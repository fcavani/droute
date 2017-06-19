// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package bucket

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/fcavani/droute/responsewriter"
)

func TestLeekingBucket(t *testing.T) {
	handler := NewBucket(1, time.Second, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Fprint(rw, "oi")
	}))

	rw := responsewriter.NewResponseWriter()
	req, err := http.NewRequest("GET", "https://dummy/", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(rw, req)

	if string(rw.Bytes()) != "oi" {
		t.Fatal("wrong anwser")
	}
}

func TestLeekingBucketTimeout(t *testing.T) {
	handler := NewBucket(1, time.Millisecond, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(10 * time.Second) //BUG: from 1ms to 10s is a lot of time
		fmt.Fprint(rw, "oi")
	}))

	rw := responsewriter.NewResponseWriter()
	req, err := http.NewRequest("GET", "https://dummy/", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(rw, req)

	if rw.ResponseCode() != http.StatusServiceUnavailable {
		t.Fatal("wrong response code", rw.ResponseCode())
	}
}
