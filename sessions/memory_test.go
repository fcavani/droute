// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package sessions

import (
	"testing"
	"time"

	"github.com/fcavani/e"
)

func TestMemEntry(t *testing.T) {
	session := NewMemEntry("uuid_01")
	if session.UUID() != "uuid_01" {
		t.Fatal("uuid fail")
	}
	err := session.Set("key_a", "a")
	if err != nil {
		t.Fatal(err)
	}
	err = session.Set("key_a", "a")
	if err != nil && !e.Equal(err, ErrSessionDataKeyExist) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err nil")
	}
	data, err := session.Get("key_a")
	if err != nil {
		t.Fatal(err)
	}
	if d := data.(string); d != "a" {
		t.Fatal("wrong data")
	}
	_, err = session.Get("blá")
	if err != nil && !e.Equal(err, ErrSessionDataNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err nil")
	}
	err = session.Del("blá")
	if err != nil && !e.Equal(err, ErrSessionDataNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err nil")
	}
	err = session.Del("key_a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = session.Get("key_a")
	if err != nil && !e.Equal(err, ErrSessionDataNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err nil")
	}
}

func TestSession(t *testing.T) {
	now := time.Date(2000, 12, 12, 12, 12, 12, 12, time.UTC)
	s := &session{
		Session:  NewMemEntry("uuid_01"),
		lastSeen: now,
	}

	if !s.KeyField().(time.Time).Equal(now) {
		t.Fatal("not equal")
	}

	if !s.Equal(now) {
		t.Fatal("equal fail")
	}

	if s.Equal(time.Now()) {
		t.Fatal("not equal fail")
	}

	if !s.Less(time.Date(2000, 12, 12, 12, 12, 12, 13, time.UTC)) {
		t.Fatal("not less")
	}
	if s.Less(time.Date(2000, 12, 12, 12, 12, 12, 11, time.UTC)) {
		t.Fatal("less")
	}
}

var sessions Sessions

func TestMemSessions(t *testing.T) {
	sessions = NewMemSessions(60*time.Second, 20*time.Second)
}

func TestMemSessionsNew(t *testing.T) {
	sessionsNew(t, sessions)
}

func TestMemSessionsRestore(t *testing.T) {
	sessionsRestore(t, sessions)
}

func TestMemSessionsDelete2(t *testing.T) {
	sessionsDelete(t, sessions)
}

func TestMemSessionsDelete(t *testing.T) {
	err := sessions.Delete("zzzzzz....")
	if err != nil && !e.Equal(err, ErrSessNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err nil")
	}

	// err = sessions.Delete("abcd")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	s, err := sessions.New("abcd")
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan struct{})

	go func() {
		err := sessions.Delete("abcd")
		if err != nil {
			t.Fatal(err)
		}
		ch <- struct{}{}
	}()

	err = sessions.Return(s)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-ch:
		return
	case <-time.After(time.Millisecond * 100):
		t.Fatal("session not deleted")
	}
}

// func TestMemSessionsDel(t *testing.T) {
// 	tx := NewTransaction()
//
// 	s, err := sessions.New("abcd")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	ch := make(chan struct{})
//
// 	go func() {
// 		err := sessions.(*MemSessions).del(tx, "abcd")
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		ch <- struct{}{}
// 	}()
//
// 	err = sessions.Return(s)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	select {
// 	case <-ch:
// 		o, err := tx.Commit(sessions.(*MemSessions).o)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		if len(o) > 0 {
// 			t.Fatal("not deleted")
// 		}
// 	case <-time.After(time.Millisecond * 100):
// 		t.Fatal("session not deleted")
// 	}
// }

func TestMemSessionsInUse(t *testing.T) {
	sessionsInUse(t, sessions)
}

func TestMemSessionsTtl(t *testing.T) {
	sessionsTtl(t, sessions)
}

func TestSessionsClose(t *testing.T) {
	err := sessions.Close()
	if err != nil {
		t.Fatal(err)
	}
}
