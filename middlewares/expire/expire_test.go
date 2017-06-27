// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package expire

import (
	"net/http"
	"testing"
	"time"

	"github.com/fcavani/droute/responsewriter"
)

func TestExpireZero(t *testing.T) {
	w := responsewriter.NewResponseWriter()
	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	Expire(0, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})(w, r)
	if w.Header().Get("Expires") != "" {
		t.Fatal("unfortunately the expire header was set")
	}
}

func TestExpire(t *testing.T) {
	w := responsewriter.NewResponseWriter()
	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	expire := 5 * time.Millisecond
	start := time.Now().Add(expire)
	Expire(expire, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})(w, r)
	stop := time.Now().Add(expire)
	if w.Header().Get("Expires") == "" {
		t.Fatal("expires empty")
	}
	exp, err := time.Parse(time.RFC1123, w.Header().Get("Expires"))
	if err != nil {
		t.Fatal(err)
	}
	if !exp.After(start) && !exp.Before(stop) {
		t.Fatal("expire set to wrong date")
	}
}
