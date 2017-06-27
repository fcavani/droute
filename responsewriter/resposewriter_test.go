// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package responsewriter

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func TestResponseWriter(t *testing.T) {
	str := "catotos"
	rw := NewResponseWriter()
	n, err := rw.Write([]byte(str))
	if err != nil {
		t.Fatal(err)
	}
	if n != len(str) {
		t.Fatal("didn´t write all the string")
	}
	rw.WriteHeader(200)
	if code := rw.ResponseCode(); code != 200 {
		t.Fatal("wrong response code", code)
	}
	rw.Header().Add("foo", "bar")
	if v := rw.Header().Get("foo"); v != "bar" {
		t.Fatal("header didn´t work, wrong value", v)
	}
	if rw.Len() != len(str) {
		t.Fatal("invalid length")
	}
	if buf := rw.Bytes(); string(buf) != str {
		t.Fatal("buffer is invalid")
	}

	rw.Reset()
	if rw.Len() != 0 {
		t.Fatal("invalid length")
	}
	if code := rw.ResponseCode(); code != 0 {
		t.Fatal("wrong response code", code)
	}
	_, err = rw.Write([]byte(str))
	if err != nil {
		t.Fatal(err)
	}
	rw.WriteHeader(200)
	rw.Header().Add("foo", "bar")

	dst := NewResponseWriter()
	err = rw.Copy(dst)
	if err != nil {
		t.Fatal(err)
	}
	if code := dst.ResponseCode(); code != 200 {
		t.Fatal("wrong response code", code)
	}
	if v := dst.Header().Get("foo"); v != "bar" {
		t.Fatal("header didn´t work, wrong value", v)
	}
	if dst.Len() != len(str) {
		t.Fatal("invalid length")
	}
	buf, err := ioutil.ReadAll(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf) != str {
		t.Fatal("read failed", string(buf))
	}
}

func TestHandler(t *testing.T) {
	str := "fullbuffer"
	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	rw := NewResponseWriter()
	h := Handler(func(rw *ResponseWriter, r *http.Request) {
		rw.WriteHeader(200)
		rw.Write([]byte(str))
	})
	h(rw, r)
	buf, err := ioutil.ReadAll(rw)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf) != str {
		t.Fatal("read failed", string(buf))
	}
	if code := rw.ResponseCode(); code != 200 {
		t.Fatal("wrong code", code)
	}
}
