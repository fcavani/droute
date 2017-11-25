// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package sessions

import (
	"testing"

	"github.com/fcavani/e"
)

func sessionsNew(t *testing.T, s Sessions) {
	sess, err := s.New("aaa")
	if err != nil {
		t.Fatal(err)
	}
	err = sess.Set("foo", "bar")
	if err != nil {
		t.Fatal(err)
	}
	dat, err := sess.Get("foo")
	if err != nil {
		t.Fatal(err)
	}
	if dat.(string) != "bar" {
		t.Fatal("retriver wrong data")
	}
	_, err = sess.Get("nodata")
	if err != nil && !e.Equal(err, ErrSessionDataNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err is nil")
	}
	err = sess.Del("nodata")
	if err != nil && !e.Equal(err, ErrSessionDataNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err is nil")
	}
	err = sess.Del("foo")
	if err != nil {
		t.Fatal(err)
	}
	err = sess.Set("foo", "bar")
	if err != nil {
		t.Fatal(err)
	}
	err = s.Return(sess)
	if err != nil {
		t.Fatal(err)
	}
}

func sessionsRestore(t *testing.T, s Sessions) {
	sess, err := s.Restore("aaa")
	if err != nil {
		t.Fatal(err)
	}
	dat, err := sess.Get("foo")
	if err != nil {
		t.Fatal(err)
	}
	if dat.(string) != "bar" {
		t.Fatal("retriver wrong data")
	}
	err = sess.Del("foo")
	if err != nil {
		t.Fatal(err)
	}
	err = s.Return(sess)
	if err != nil {
		t.Fatal(err)
	}
}

func sessionsDelete(t *testing.T, s Sessions) {
	sess, err := s.Restore("aaa")
	if err != nil {
		t.Fatal(err)
	}
	err = s.Delete("aaa")
	if err != nil && !e.Equal(err, ErrSessInUse) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err is nil")
	}
	err = s.Delete("foo")
	if err != nil && !e.Equal(err, ErrSessNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err is nil")
	}
	err = s.Return(sess)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Delete("aaa")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Restore("aaa")
	if err != nil && !e.Equal(err, ErrSessNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err is nil")
	}
}

func sessionsInUse(t *testing.T, s Sessions) {
	_, err := s.IsInUse("aaa")
	if err != nil && !e.Equal(err, ErrSessNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err is nil")
	}
	sess, err := s.New("aaa")
	if err != nil {
		t.Fatal(err)
	}
	inuse, err := s.IsInUse("aaa")
	if err != nil {
		t.Fatal(err)
	}
	if !inuse {
		t.Fatal("not in use")
	}
	err = s.Return(sess)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Delete("aaa")
	if err != nil {
		t.Fatal(err)
	}
}

func sessionsTtl(t *testing.T, s Sessions) {
	_, err := s.Ttl("zzzz")
	if err != nil && !e.Equal(err, ErrSessNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err is nil")
	}
	sess, err := s.New("aaa")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Ttl("aaa")
	if err != nil {
		t.Fatal(err)
	}
	err = s.Return(sess)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Delete("aaa")
	if err != nil {
		t.Fatal(err)
	}
}
