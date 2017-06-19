// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/fcavani/e"
)

func userURL(req *http.Request) (*url.URL, error) {
	scheme := "http"
	if req.TLS != nil {
		scheme += "s"
	}
	host := req.Host
	if h := req.Header.Get("X-Forwarded-Host"); h != "" {
		host = h
	}
	query := ""
	if req.URL.RawQuery != "" {
		query = "?" + req.URL.RawQuery
	}
	fragment := ""
	if req.URL.Fragment != "" {
		fragment = "#" + req.URL.Fragment
	}
	path := ""
	s := strings.Split(req.RequestURI, "?")
	if len(s) >= 1 {
		path = s[0]
	}
	//path = root + path
	uri := scheme + "://" + host + path + query + fragment
	myURL, err := url.Parse(uri)
	if err != nil {
		return nil, e.New(err)
	}
	myURL.Path = path
	return myURL, nil
}
