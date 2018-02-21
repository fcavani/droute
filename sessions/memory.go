// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package sessions

import (
	"sync"
	"time"

	"github.com/fcavani/e"
	log "github.com/fcavani/slog"
)

type MemEntry struct {
	ID    string
	Store map[string]interface{}
	lck   sync.RWMutex
}

func NewMemEntry(uuid string) Session {
	return &MemEntry{
		ID:    uuid,
		Store: make(map[string]interface{}),
	}
}

func (m *MemEntry) UUID() string {
	return m.ID
}

func (m *MemEntry) Get(key string) (interface{}, error) {
	m.lck.RLock()
	defer m.lck.RUnlock()
	data, found := m.Store[key]
	if !found {
		return nil, e.New(ErrSessionDataNotFound)
	}
	return data, nil
}

func (m *MemEntry) Set(key string, data interface{}) error {
	m.lck.Lock()
	defer m.lck.Unlock()
	_, found := m.Store[key]
	if found {
		return e.New(ErrSessionDataKeyExist)
	}
	m.Store[key] = data
	return nil
}

func (m *MemEntry) Del(key string) error {
	m.lck.Lock()
	defer m.lck.Unlock()
	_, found := m.Store[key]
	if !found {
		return e.New(ErrSessionDataNotFound)
	}
	delete(m.Store, key)
	return nil
}

type session struct {
	Session
	lastSeen time.Time
}

func (s *session) KeyField() interface{} {
	return s.lastSeen
}

func (s *session) Less(x interface{}) bool {
	data := x.(time.Time)
	return s.lastSeen.Before(data)
}

func (s *session) Equal(x interface{}) bool {
	data := x.(time.Time)
	return s.lastSeen.Equal(data)
}

type MemSessions struct {
	sessions map[string]*session
	o        Ordered
	lck      sync.Mutex
	ttl      time.Duration
	chclose  chan chan struct{}
	count    map[string]uint
	signal   map[string][]chan struct{}
	isclosed bool
	dsignal  chan uint
}

func NewMemSessions(ttl, cleanup time.Duration) Sessions {
	m := &MemSessions{
		sessions: make(map[string]*session),
		o:        make(Ordered, 0),
		ttl:      ttl,
		chclose:  make(chan chan struct{}),
		count:    make(map[string]uint),
		signal:   make(map[string][]chan struct{}),
	}
	go func() {
		var err error
		for {
			select {
			case <-time.After(cleanup):
				err = m.o.Iter(func(i int, data interface{}, tx *Transaction) error {
					s := data.(*session)
					if time.Now().Before(s.lastSeen) {
						err = m.del(tx, s.UUID())
						if err != nil {
							log.Tag("session", "cleanup", "del").Errorln("Fail to delete session:", err)
							return nil
						}
					}
					return nil
				})
				if err != nil {
					log.Tag("session", "cleanup", "del").Errorln("Fail to iter sessions:", err)
				}
			case ch := <-m.chclose:
				ch <- struct{}{}
			}
		}
	}()
	return m
}

func (m *MemSessions) Close() error {
	m.lck.Lock()
	if m.isclosed {
		m.lck.Unlock()
		return nil
	}
	m.lck.Unlock()

	ch := make(chan struct{})
	m.chclose <- ch
	<-ch

	// Em algum lugar aqui eu tenho que desabilitar os outros métodos.

again:
	m.lck.Lock()

	m.isclosed = true

	if !m.sessionAny() {
		//No sessions to return
		m.dsignal = nil
		m.lck.Unlock()
		return nil
	}

	m.dsignal = make(chan uint)
	m.lck.Unlock()
	<-m.dsignal
	goto again

	return nil
}

func (m *MemSessions) sessionInc(uuid string) {
	if _, found := m.count[uuid]; !found {
		m.count[uuid] = 1
		return
	}
	m.count[uuid] = m.count[uuid] + 1
}

func (m *MemSessions) sessionInUse(uuid string) bool {
	if _, found := m.count[uuid]; !found {
		return false
	}
	return m.count[uuid] > 0
}

func (m *MemSessions) sessionDec(uuid string) {
	if _, found := m.count[uuid]; !found {
		m.count[uuid] = 0
		return
	}
	defer func() {
		if m.dsignal != nil {
			m.dsignal <- m.count[uuid]
		}
	}()
	if m.count[uuid] <= 0 {
		return
	}
	m.count[uuid] = m.count[uuid] - 1
}

func (m *MemSessions) sessionCount(uuid string) uint {
	if _, found := m.count[uuid]; !found {
		return 0
	}
	return m.count[uuid]
}

func (m *MemSessions) sessionAny() bool {
	for _, v := range m.count {
		if v > 0 {
			return true
		}
	}
	return false
}

func (m *MemSessions) New(uuid string) (Session, error) {
	m.lck.Lock()
	defer m.lck.Unlock()

	if _, found := m.sessions[uuid]; found {
		return nil, e.Forward(ErrSessFound)
	}

	s := &session{
		Session:  NewMemEntry(uuid),
		lastSeen: time.Now().Add(m.ttl),
	}

	err := m.o.Set(s)
	if err != nil {
		return nil, e.Forward(err)
	}
	m.sessions[uuid] = s

	m.sessionInc(uuid)

	return s, nil
}

func (m *MemSessions) Restore(uuid string) (Session, error) {
	m.lck.Lock()
	defer m.lck.Unlock()

	s, found := m.sessions[uuid]
	if !found {
		return nil, e.New(ErrSessNotFound)
	}

	err := m.o.Del(s)
	if err != nil {
		return nil, e.Forward(err)
	}
	s.lastSeen = time.Now().Add(m.ttl)
	err = m.o.Set(s)
	if err != nil {
		return nil, e.Forward(err)
	}

	m.sessionInc(uuid)

	return s, nil
}

func (m *MemSessions) Delete(uuid string) error {
	// Problema do segundo Delete.
	// Não é problema... o segundo vai falhar.
	m.lck.Lock()
	defer m.lck.Unlock()

	s, found := m.sessions[uuid]
	if !found {
		return e.New(ErrSessNotFound)
	}

	if count := m.sessionCount(uuid); count > 0 {
		//wait for the session be give back
		// ch := make(chan struct{})
		// m.signal[uuid] = append(m.signal[uuid], ch)
		// m.lck.Unlock()
		// <-ch
		// m.lck.Lock()
		return e.New(ErrSessInUse)
	}

	err := m.o.Del(s)
	if err != nil {
		return e.Forward(err)
	}

	delete(m.sessions, uuid)

	return nil
}

func (m *MemSessions) del(tx *Transaction, uuid string) error {
	// Problema do segundo Delete.
	// Não é problema... o segundo vai falhar.
	m.lck.Lock()
	defer m.lck.Unlock()

	s, found := m.sessions[uuid]
	if !found {
		return e.New(ErrSessNotFound)
	}

	if count := m.sessionCount(uuid); count > 0 {
		//wait for the session be give back
		// ch := make(chan struct{})
		// m.signal[uuid] = append(m.signal[uuid], ch)
		// m.lck.Unlock()
		// <-ch
		// m.lck.Lock()
		return e.New(ErrSessInUse)
	}

	err := tx.Del(s)
	if err != nil {
		return e.Forward(err)
	}
	delete(m.sessions, uuid)

	return nil
}

func (m *MemSessions) Return(s Session) error {
	m.lck.Lock()
	defer m.lck.Unlock()

	_, found := m.sessions[s.UUID()]
	if !found {
		return e.New(ErrSessNotFound)
	}

	defer m.sessionDec(s.UUID())

	chs, found := m.signal[s.UUID()]
	if !found {
		return nil
	}

	for _, ch := range chs {
		ch <- struct{}{}
	}

	delete(m.signal, s.UUID())

	return nil
}

func (m *MemSessions) IsInUse(uuid string) (bool, error) {
	m.lck.Lock()
	defer m.lck.Unlock()

	_, found := m.sessions[uuid]
	if !found {
		return false, e.New(ErrSessNotFound)
	}

	return m.sessionInUse(uuid), nil
}

func (m *MemSessions) Ttl(uuid string) (time.Time, error) {
	m.lck.Lock()
	defer m.lck.Unlock()

	s, found := m.sessions[uuid]
	if !found {
		return time.Time{}, e.New(ErrSessNotFound)
	}

	if s.lastSeen.Before(time.Now()) {
		return time.Time{}, e.New(ErrSessExpired)
	}

	return s.lastSeen, nil
}
