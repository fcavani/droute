// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/fcavani/droute/middlewares/bucket"
	"github.com/fcavani/droute/responsewriter"
	"github.com/fcavani/e"
	"gopkg.in/fcavani/httprouter.v2"
)

func TestRouters(t *testing.T) {
	r := NewRouters()
	tr := httprouter.New()
	r.Set("teste", tr)
	def := r.Get(DefaultRouter)
	teste := r.Get("teste")
	if teste != tr {
		t.Fatal("get failed")
	}
	foo := r.Get("bar")
	if foo != def {
		t.Fatal("get default failed")
	}
	r.Del("teste")
	teste = r.Get("teste")
	if teste != def {
		t.Fatal("get failed")
	}
	r.Del(DefaultRouter)
	def = r.Get(DefaultRouter)
	if def != nil {
		t.Fatal("get nothing failed")
	}
}

func TestRouter(t *testing.T) {
	routers := NewRouters()
	lb := NewRoundRobin()
	r := &Router{}

	err := r.Start(routers, lb, 60*time.Second, 5)
	if err != nil {
		t.Fatal(err)
	}

	r.HTTPHandlers(func(first http.Handler) http.Handler {
		return bucket.NewBucket(
			10,
			60000*time.Millisecond,
			first,
		)
	})

	r.Middlewares(func(last responsewriter.HandlerFunc) responsewriter.HandlerFunc {
		return last
	})

	r.SetHTTPAddr("foo")
	r.SetHTTPSAddr("foo")

	w := responsewriter.NewResponseWriter()
	req, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	r.ServeHTTP(w, req)

	if code := w.ResponseCode(); code != 404 {
		t.Fatal("wrong response code", code)
	}
}

func TestRouter2(t *testing.T) {
	r := &Router{}

	err := r.Start(NewRouters(), NewRoundRobin(), 60*time.Second, 3)
	if err != nil {
		t.Fatal(err)
	}

	HTTPClient = &http.Client{
		Transport: &transport{},
	}
	defer func() {
		HTTPClient = http.DefaultClient
	}()

	err = r.Add(DefaultRouter, "GET", "", "10.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	w := responsewriter.NewResponseWriter()
	req, err := http.NewRequest("GET", "http://localhost/", bytes.NewBufferString("oi"))
	if err != nil {
		t.Fatal(err)
	}

	r.ServeHTTP(w, req)

	if code := w.ResponseCode(); code != 200 {
		t.Fatal("wrong response code", code)
	}
	if dst := w.Header().Get("X-Dst-Serv"); dst != "10.0.0.1" {
		t.Fatal("wrong destiny", dst)
	}
	if buf := w.Bytes(); string(buf) != "oi" {
		t.Fatal("response error", string(buf))
	}
}

func TestRouterNoDefRoute(t *testing.T) {
	rs := NewRouters()
	lb := NewRoundRobin()
	r := &Router{}

	rs.Del(DefaultRouter)

	err := r.Start(rs, lb, 60*time.Second, 5)
	if err != nil && !e.Contains(err, "no default router") {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error is nil")
	}
}

func TestRouterAdd(t *testing.T) {
	r := &Router{}

	err := r.Start(NewRouters(), NewRoundRobin(), 60*time.Second, 3)
	if err != nil {
		t.Fatal(err)
	}

	HTTPClient = &http.Client{
		Transport: &transport{},
	}
	defer func() {
		HTTPClient = http.DefaultClient
	}()

	err = r.Add("a", "GET", "", "10.0.0.1")
	if err != nil && !e.Contains(err, "invalid route name") {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("nil error")
	}

	err = r.Add(DefaultRouter, "?", "", "10.0.0.1")
	if err != nil && !e.Contains(err, "invalid method name") {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("nil error")
	}

	err = r.Add(DefaultRouter, "GET", strings.Repeat("a", 256), "10.0.0.1")
	if err != nil && !e.Contains(err, "invalid path name") {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("nil error")
	}

	err = r.Add(DefaultRouter, "GET", "/", "http\n://localhost")
	if err != nil && !e.Contains(err, "invalid destiny host name") {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("nil error")
	}

	err = r.Add(DefaultRouter, "GET", "/", "10.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	// err = r.Add(DefaultRouter, "GET", "/", "10.0.0.1")
	// if err != nil && !e.Contains(err, "already registered") {
	// 	t.Fatal(err)
	// } else if err == nil {
	// 	t.Fatal("nil error")
	// }
	err = r.Add(DefaultRouter, "GET", "/", "10.0.0.2")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRouterAddRouteRouteName(t *testing.T) {
	defer func() {
		r := recover()
		if s, ok := r.(string); ok {
			if s != "invalid route name" {
				t.Fatal(s)
			}
		}
	}()

	r := &Router{}

	err := r.Start(NewRouters(), NewRoundRobin(), 60*time.Second, 3)
	if err != nil {
		t.Fatal(err)
	}

	r.AddRoute("a", "GET", "", nil)
}

func TestRouterAddRouteMethod(t *testing.T) {
	defer func() {
		r := recover()
		if s, ok := r.(string); ok {
			if s != "invalid method name" {
				t.Fatal(s)
			}
		}
	}()

	r := &Router{}

	err := r.Start(NewRouters(), NewRoundRobin(), 60*time.Second, 3)
	if err != nil {
		t.Fatal(err)
	}

	r.AddRoute(DefaultRouter, "a", "", nil)
}

func TestRouterAddRoutePath(t *testing.T) {
	defer func() {
		r := recover()
		if s, ok := r.(string); ok {
			if s != "invalid path name" {
				t.Fatal(s)
			}
		}
	}()

	r := &Router{}

	err := r.Start(NewRouters(), NewRoundRobin(), 60*time.Second, 3)
	if err != nil {
		t.Fatal(err)
	}

	r.AddRoute(DefaultRouter, "GET", "", nil)
	r.AddRoute(DefaultRouter, "GET", strings.Repeat("a", 129), nil)
}

func TestRouterAddRouteRoute(t *testing.T) {
	defer func() {
		r := recover()
		if s, ok := r.(string); ok {
			if s != "route not found" {
				t.Fatal(s)
			}
		}
	}()

	r := &Router{}
	routers := NewRouters()

	err := r.Start(routers, NewRoundRobin(), 60*time.Second, 3)
	if err != nil {
		t.Fatal(err)
	}

	routers.Del(DefaultRouter)

	r.AddRoute("foobar", "GET", "", nil)
}

func TestRouterAddHandler(t *testing.T) {
	r := &Router{}

	err := r.Start(NewRouters(), NewRoundRobin(), 60*time.Second, 3)
	if err != nil {
		t.Fatal(err)
	}

	HTTPClient = &http.Client{
		Transport: &transport{},
	}
	defer func() {
		HTTPClient = http.DefaultClient
	}()

	buf, err := json.Marshal(&Route{
		Methode: "GET",
		Router:  DefaultRouter,
		Path:    "/",
		RedirTo: "10.0.0.1",
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "http://localhost/_router/add", bytes.NewBuffer(buf))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("X-Real-Ip", "127.0.0.1")
	rw := responsewriter.NewResponseWriter()
	r.ServeHTTP(rw, req)
	if code := rw.ResponseCode(); code != 201 {
		t.Fatal("wrong response code", code)
	}
	var resp Response
	err = json.NewDecoder(rw).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != "" {
		t.Fatal(resp.Err)
	}
	if resp.Error() != "no error" {
		t.Fatal(resp.Error())
	}
	if resp.Methode != "GET" || resp.Path != "/" || resp.Router != DefaultRouter {
		t.Fatal("wrong response")
	}
	if resp.Op != "add" {
		t.Fatal("wrong operation", resp.Op)
	}

	req, err = http.NewRequest("POST", "http://localhost/_router/add", bytes.NewBuffer(buf))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("X-Real-Ip", "10.0.0.1")
	rw = responsewriter.NewResponseWriter()
	r.ServeHTTP(rw, req)
	if code := rw.ResponseCode(); code != 403 {
		t.Fatal("wrong response code", code)
	}

	req, err = http.NewRequest("POST", "http://localhost/_router/add", bytes.NewBuffer(buf))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("X-Real-Ip", "localhost")
	rw = responsewriter.NewResponseWriter()
	r.ServeHTTP(rw, req)
	if code := rw.ResponseCode(); code != 403 {
		t.Fatal("wrong response code", code)
	}

	req, err = http.NewRequest("POST", "http://localhost/_router/add", bytes.NewBufferString(strings.Repeat("a", int(BodyLimitSize)+1)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("X-Real-Ip", "127.0.0.1")
	rw = responsewriter.NewResponseWriter()
	r.ServeHTTP(rw, req)
	if code := rw.ResponseCode(); code != 500 {
		t.Fatal("wrong response code", code)
	}
	var operr OpErr
	err = json.NewDecoder(rw).Decode(&operr)
	if err != nil {
		t.Fatal(err)
	}
	if operr.Err != "invalid character 'a' looking for beginning of value" {
		t.Fatal("wrong error:", operr.Err)
	}
	if operr.Op != "add" {
		t.Fatal("wrong op", operr.Op)
	}
	if operr.Error() != "op: add, err: invalid character 'a' looking for beginning of value" {
		t.Fatal("wrong error", operr.Error())
	}

	buf, err = json.Marshal(&Route{
		Methode: "a",
		Router:  DefaultRouter,
		Path:    "/",
		RedirTo: "10.0.0.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err = http.NewRequest("POST", "http://localhost/_router/add", bytes.NewBuffer(buf))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("X-Real-Ip", "127.0.0.1")
	rw = responsewriter.NewResponseWriter()
	r.ServeHTTP(rw, req)
	if code := rw.ResponseCode(); code != 422 {
		t.Fatal("wrong response code", code)
	}
	err = json.NewDecoder(rw).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.Error())
	if !strings.Contains(resp.Err, "invalid method name") {
		t.Fatal("wrong error", resp.Err)
	}
}
