// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/fcavani/droute/responsewriter"
)

type transport struct {
	Err     error
	Timeout time.Duration
}

func (trans *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if trans.Err != nil {
		return nil, trans.Err
	}
	if trans.Timeout != 0 {
		time.Sleep(trans.Timeout)
	}
	close := false
	if req.Body == nil {
		close = true
	}
	resp := &http.Response{
		Status:           "200 OK",
		StatusCode:       200,
		Proto:            "HTTP/1.0",
		ProtoMajor:       1,
		ProtoMinor:       0,
		Header:           req.Header,
		Body:             req.Body,
		ContentLength:    req.ContentLength,
		TransferEncoding: nil,
		Close:            close,
		Uncompressed:     true,
		Trailer:          make(http.Header),
		Request:          req,
		TLS:              nil,
	}
	return resp, nil
}

func TestProxy(t *testing.T) {
	HTTPClient = &http.Client{
		Transport: &transport{},
	}
	rd := NewRedirDst("10.0.0.1")
	h := Balance(rd, Proxy("", 300*time.Millisecond))

	w := responsewriter.NewResponseWriter()
	r, err := http.NewRequest("GET", "http://blurft", nil)
	if err != nil {
		t.Fatal(err)
	}
	h(w, r)
	if code := w.ResponseCode(); code != 200 {
		t.Fatal("response code is wrong", code)
	}
	if dst := w.Header().Get("X-Dst-Serv"); dst != "10.0.0.1" {
		t.Fatal("wrong destiny", dst)
	}

	w = responsewriter.NewResponseWriter()
	r, err = http.NewRequest("GET", "http://blurft", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.URL.Scheme = "ht\n://"
	h(w, r)
	if code := w.ResponseCode(); code != 500 {
		t.Fatal("response code is wrong", code)
	}

	w = responsewriter.NewResponseWriter()
	buf := bytes.NewBufferString("oi")
	r, err = http.NewRequest("GET", "http://blurft", buf)
	if err != nil {
		t.Fatal(err)
	}
	h(w, r)
	if code := w.ResponseCode(); code != 200 {
		t.Fatal("response code is wrong", code)
	}
	if buf := w.Bytes(); string(buf) != "oi" {
		t.Fatal("response error")
	}

	w = responsewriter.NewResponseWriter()
	buf = bytes.NewBufferString("oi")
	r, err = http.NewRequest("GET", "http://blurft", buf)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("foo", "bar")
	h(w, r)
	if code := w.ResponseCode(); code != 200 {
		t.Fatal("response code is wrong", code)
	}
	if buf := w.Bytes(); string(buf) != "oi" {
		t.Fatal("response error")
	}
	if hdr := w.Header().Get("foo"); hdr != "bar" {
		t.Fatal("header error", hdr)
	}

	w = responsewriter.NewResponseWriter()
	ebuf := &errorBuf{}
	r, err = http.NewRequest("GET", "http://blurft", ebuf)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("foo", "bar")
	h(w, r)
	if code := w.ResponseCode(); code != 500 {
		t.Fatal("response code is wrong", code)
	}

	HTTPClient = &http.Client{
		Transport: &transport{
			Err: errors.New("dummy error"),
		},
	}
	w = responsewriter.NewResponseWriter()
	r, err = http.NewRequest("GET", "http://blurft", buf)
	if err != nil {
		t.Fatal(err)
	}
	h(w, r)
	if code := w.ResponseCode(); code != 500 {
		t.Fatal("response code is wrong", code)
	}
}

func TestDestiny(t *testing.T) {
	HTTPClient = &http.Client{
		Transport: &transport{},
	}

	h := emptyDst(Proxy("", 300*time.Millisecond))

	w := responsewriter.NewResponseWriter()
	r, err := http.NewRequest("GET", "http://blurft", nil)
	if err != nil {
		t.Fatal(err)
	}
	h(w, r)
	if code := w.ResponseCode(); code != 500 {
		t.Fatal("response code is wrong", code)
	}
}

func TestTimeout(t *testing.T) {
	HTTPClient = &http.Client{
		Transport: &transport{
			Timeout: 600 * time.Millisecond,
		},
	}
	rd := NewRedirDst("10.0.0.1")
	h := Balance(rd, Proxy("", 300*time.Millisecond))

	w := responsewriter.NewResponseWriter()
	r, err := http.NewRequest("GET", "http://blurft", nil)
	if err != nil {
		t.Fatal(err)
	}
	h(w, r)
	if code := w.ResponseCode(); code != 408 {
		t.Fatal("response code is wrong", code)
	}
}

type errorBuf struct{}

func (eb *errorBuf) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func (eb *errorBuf) Close() error {
	return nil
}

func emptyDst(handler responsewriter.HandlerFunc) responsewriter.HandlerFunc {
	return func(rw *responsewriter.ResponseWriter, req *http.Request) {
		req = req.WithContext(context.WithValue(req.Context(), "proxyredirdst", "\n://bugbug"))
		handler(rw, req)
	}
}
