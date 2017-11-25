// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package sessions

import (
	"sync"

	"github.com/fcavani/e"
)

type CacheEntry struct {
	Sess  Session
	Store map[string]interface{}
	ID    string
	lck   sync.RWMutex
}

func NewCachedEntry(s Session) Session {
	return &CacheEntry{
		Sess:  s,
		Store: make(map[string]interface{}),
		ID:    s.UUID(),
	}
}

func (c *CacheEntry) UUID() string {
	return c.ID
}

func (c *CacheEntry) Get(key string) (interface{}, error) {
	var err error
	c.lck.RLock()
	data, found := c.Store[key]
	if found {
		c.lck.RUnlock()
		return data, nil
	}
	c.lck.RUnlock()
	data, err = c.Sess.Get(key)
	if err != nil {
		return nil, e.Forward(err)
	}
	c.lck.Lock()
	c.Store[key] = data
	c.lck.Unlock()
	return data, nil
}

func (c *CacheEntry) Set(key string, data interface{}) error {
	err := c.Sess.Set(key, data)
	if err != nil {
		return e.Forward(err)
	}
	c.lck.Lock()
	c.Store[key] = data
	c.lck.Unlock()
	return nil
}

func (c *CacheEntry) Del(key string) error {
	c.lck.Lock()
	delete(c.Store, key)
	c.lck.Unlock()
	err := c.Sess.Del(key)
	if err != nil {
		return e.Forward(err)
	}
	return nil
}
