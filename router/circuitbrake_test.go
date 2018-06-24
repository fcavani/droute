// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"net/http"
	"testing"

	"github.com/fcavani/droute/responsewriter"
)

func TestBrake(t *testing.T) {
	dst := "10.0.1.1"
	rr := NewRoundRobin()
	rr.AddAddrs("GET", "/", dst)
	cbs := make(Cbs)
	rw := responsewriter.NewResponseWriter()
	r, err := http.NewRequest("GET", "http://localhost/en", nil)
	if err != nil {
		t.Fatal(err)
	}

	h := Balance(rr, CircuitBrake(cbs, func(rw *responsewriter.ResponseWriter, r *http.Request) {
		proxy := r.Context().Value("proxyredirdst").(string)
		t.Log(proxy)
		rw.WriteHeader(200)
	}))
	h(rw, r)
	if code := rw.ResponseCode(); code != 200 {
		t.Fatal("wrong response code", code)
	}
	cb, found := cbs[dst]
	if !found {
		t.Fatal("circuit brake for this server not found", dst)
	}
	s := cb.State()
	t.Log(s)

	h = Balance(rr, CircuitBrake(cbs, func(rw *responsewriter.ResponseWriter, r *http.Request) {
		proxy := r.Context().Value("proxyredirdst").(string)
		t.Log(proxy)
		rw.WriteHeader(500)
	}))
	h(rw, r)
	if code := rw.ResponseCode(); code != 500 {
		t.Fatal("wrong response code", code)
	}
	_, found = cbs[dst]
	if found {
		t.Fatal("circuit brake for this server found", dst)
	}
}
