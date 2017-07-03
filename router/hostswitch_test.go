// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"net/http"
	"testing"

	"github.com/fcavani/droute/responsewriter"
)

func TestHostSwitch(t *testing.T) {
	hs := make(HostSwitch)
	hs.Set("localhost", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(200)
		rw.Write([]byte("oi"))
	}))

	req, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := responsewriter.NewResponseWriter()
	hs.ServeHTTP(w, req)
	if code := w.ResponseCode(); code != 200 {
		t.Fatal("wrong code", code)
	}
	if str := string(w.Bytes()); str != "oi" {
		t.Fatal("wrong response", str)
	}

	req, err = http.NewRequest("GET", "http://mars", nil)
	if err != nil {
		t.Fatal(err)
	}
	w = responsewriter.NewResponseWriter()
	hs.ServeHTTP(w, req)
	if code := w.ResponseCode(); code != 403 {
		t.Fatal("wrong code", code)
	}
	if str := string(w.Bytes()); str != "Forbidden\n" {
		t.Fatal("wrong response", str)
	}

	req, err = http.NewRequest("GET", "http://mars", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "xn---"
	w = responsewriter.NewResponseWriter()
	hs.ServeHTTP(w, req)
	if code := w.ResponseCode(); code != 403 {
		t.Fatal("wrong code", code)
	}
	if str := string(w.Bytes()); str != "Invalid host name\n" {
		t.Fatal("wrong response", str)
	}
}
