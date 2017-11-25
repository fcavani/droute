// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package sessions

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"
)

var s Sessions

func TestOpenSessionsBoltDB(t *testing.T) {
	var err error
	dir, err := ioutil.TempDir("", "sessiondb_")
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(dir, "sessions.db")
	s, err = OpenSessionsBoltDB(filename, 0600, nil, 5*time.Second, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBoltDBSessionsNew(t *testing.T) {
	sessionsNew(t, s)
}

func TestBoltDBSessionsRestore(t *testing.T) {
	sessionsRestore(t, s)
}

func TestBoltDBSessionsDelete(t *testing.T) {
	sessionsDelete(t, s)
}

func TestBoltDBSessionsInUse(t *testing.T) {
	sessionsInUse(t, s)
}

func TestBoltDBSessionsTtl(t *testing.T) {
	sessionsTtl(t, s)
}

func TestBoltDBClose(t *testing.T) {
	err := s.Close()
	if err != nil {
		t.Fatal(err)
	}
}
