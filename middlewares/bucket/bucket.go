// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

// Package bucket contains a leeking bucket to be used with a web server for
// slow down the requests.
package bucket

import (
	"io"
	"net/http"
	"time"
)

type request struct {
	rw   http.ResponseWriter
	req  *http.Request
	resp chan struct{}
}

// LeekingBucket implements the bucket
type LeekingBucket struct {
	server  http.Handler
	fifo    chan *request
	timeout time.Duration
}

// NewBucket creats a new bucket
func NewBucket(size int, timeout time.Duration, s http.Handler) http.Handler {
	lb := &LeekingBucket{
		server:  s,
		fifo:    make(chan *request),
		timeout: timeout,
	}
	for i := 0; i < size; i++ {
		go func() {
			for {
				r := <-lb.fifo
				lb.server.ServeHTTP(r.rw, r.req)
				r.resp <- struct{}{}
			}
		}()
	}
	return lb
}

// ServeHTTP servers a request.
func (l *LeekingBucket) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	resp := make(chan struct{})
	l.fifo <- &request{rw: rw, req: req, resp: resp}
	select {
	case <-resp:
	case <-time.After(l.timeout):
		// Timeout error
		go func() { <-resp }()
		//log.Tag("proxy", "bucket").Println("Request timeout")
		rw.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(rw, "<html><head><title>Timeout</title></head><body><h1>Timeout</h1></body></html>")
	}
}
