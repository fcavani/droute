// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"net/http"
	"testing"

	"github.com/fcavani/droute/responsewriter"
)

func TestRedirDst(t *testing.T) {
	rd := NewRedirDst("ip")
	rd.AddAddrs("GET", "", "10.0.1.1")
	rd.Remove("GET", "", "10.0.1.1")
	if dstAddr := rd.Next("GET", "/"); dstAddr != "ip" {
		t.Fatal("invalid dst addrs", dstAddr)
	}

}

func TestRoundRobin(t *testing.T) {
	rr := NewRoundRobin()
	rr.AddAddrs("GET", "/", "10.0.1.1")
	rr.AddAddrs("GET", "/", "10.0.1.2")
	rr.AddAddrs("GET", "/", "10.0.1.3")
	for _, v := range []string{"10.0.1.1", "10.0.1.2", "10.0.1.3"} {
		dst := rr.Next("GET", "/")
		if dst != v {
			t.Fatal("addrs invalid", dst, v)
		}
	}

	dst := rr.Next("GET", "path")
	if dst != "" {
		t.Fatal("next fail")
	}
	dst = rr.Next("POST", "path")
	if dst != "" {
		t.Fatal("next fail")
	}

	rr.Remove("POST", "/", "")
	rr.Remove("GET", "/", "")
	rr.Remove("GET", "/", "10.0.1.1")
	for _, v := range []string{"10.0.1.2", "10.0.1.3"} {
		dst := rr.Next("GET", "/")
		if dst != v {
			t.Fatal("addrs invalid", dst, v)
		}
	}

	rr.AddAddrs("GET", "/*catoto", "10.0.1.1")
	rr.Remove("GET", "/*catoto", "10.0.1.1")
	rr.Remove("GET", "/*notfound", "10.0.1.1")

	rr.AddAddrs("GET", "/:foo", "10.0.1.1")
	rr.Remove("GET", "/:foo", "10.0.1.1")
	rr.Remove("GET", "/:bar", "10.0.1.1")
}

func TestHandler(t *testing.T) {
	rd := NewRedirDst("")
	rw := responsewriter.NewResponseWriter()
	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	h := Balance(rd, func(rw *responsewriter.ResponseWriter, req *http.Request) {
		rw.WriteHeader(200)
	})
	h(rw, r)

	if code := rw.ResponseCode(); code != 500 {
		t.Fatal("wrong response code", code)
	}

}

func TestMatchNamedParam(t *testing.T) {
	tests := []struct {
		path   string
		k      string
		result bool
	}{
		{"", "", false},
		{"oi", "", false},
		{"", "oi", false},
		{"oi", "oi", true},
		{"foo", "bar", false},
		{"foo", ":bar", true},
	}
	for i, test := range tests {
		if r := matchNamedParam(test.path, test.k); r != test.result {
			t.Fatal("fail", i, test.path, test.k, test.result)
		}
	}
}
