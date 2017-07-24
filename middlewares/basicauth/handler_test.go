// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package basicauth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/fcavani/droute/responsewriter"
	"github.com/fcavani/e"
)

func TestUsers(t *testing.T) {
	users := make(Users)

	err := users.SetPass("", "", "")
	if err != nil && !e.Equal(err, ErrInvName) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err is nil")
	}

	err = users.SetPass("pafuncio", "", "")
	if err != nil && !e.Equal(err, ErrBlankPass) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err is nil")
	}

	err = users.SetPass("pafuncio", "123", "")
	if err != nil && !e.Equal(err, ErrInvHash) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("err is nil")
	}

	err = users.SetPass("pafuncio", "123", "sha1")
	if err != nil {
		t.Fatal(err)
	}

	if pass, found := users["pafuncio"]; found {
		if pass != "SHA1:MTIz2jmj7l5rSw0yVb/vlWAYkK/YBwk=" {
			t.Fatal("invalid pass", pass)
		}
	} else {
		t.Fatal("user not found")
	}
}

func TestSplitHash(t *testing.T) {
	ht, pass := splitHash("SHA1:MTIz2jmj7l5rSw0yVb/vlWAYkK/YBwk=")
	if ht != "SHA1" {
		t.Fatal("invalid hash type")
	}
	if pass != "MTIz2jmj7l5rSw0yVb/vlWAYkK/YBwk=" {
		t.Fatal("invalid password hash")
	}

	ht, pass = splitHash("")
	if ht != "" || pass != "" {
		t.Fatal("invalid empty string")
	}

	ht, pass = splitHash("SHA1:")
	if ht != "" || pass != "" {
		t.Fatal("invalid string")
	}
}

func TestValidade(t *testing.T) {
	users := make(Users)
	err := users.SetPass("pafuncio", "123", "sha1")
	if err != nil {
		t.Fatal(err)
	}
	err = users.SetPass("catoto", "123", "plain")
	if err != nil {
		t.Fatal(err)
	}
	users["teste"] = ""

	if users.Validate("", "") {
		t.Fatal("invalid empty user")
	}
	if users.Validate("test", "") {
		t.Fatal("invalid pass")
	}
	if !users.Validate("catoto", "123") {
		t.Fatal("invalid user pass")
	}
	if !users.Validate("pafuncio", "123") {
		t.Fatal("invalid user pass")
	}
}

type structErr struct {
	Err string `json:"err"`
}

func TestBasicAuth(t *testing.T) {
	users := make(Users)
	err := users.SetPass("pafuncio", "123", "sha1")
	if err != nil {
		t.Fatal(err)
	}

	h := BasicAuth("helm", users, func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(200)
	})

	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	rw := responsewriter.NewResponseWriter()
	h(rw, r)
	if code := rw.ResponseCode(); code != 401 {
		t.Fatal("wrong response code", code)
	}

	r, err = http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	rw = responsewriter.NewResponseWriter()
	r.Header["Authorization"] = []string{"Basic braulho"}
	h(rw, r)
	if code := rw.ResponseCode(); code != 403 {
		t.Fatal("wrong response code", code)
	}

	r, err = http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	rw = responsewriter.NewResponseWriter()
	r.Header["Authorization"] = []string{"Basic " + base64.StdEncoding.EncodeToString([]byte("pafuncio:123"))}
	h(rw, r)
	if code := rw.ResponseCode(); code != 200 {
		t.Fatal("wrong response code", code)
	}

	r, err = http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	rw = responsewriter.NewResponseWriter()
	r.Header["Authorization"] = []string{"Basic " + base64.StdEncoding.EncodeToString([]byte("pafuncio:wrong"))}
	h(rw, r)
	if code := rw.ResponseCode(); code != 403 {
		t.Fatal("wrong response code", code)
	}
	//t.Log(string(rw.Bytes()))
	var errorResp structErr
	err = json.NewDecoder(rw).Decode(&errorResp)
	if err != nil {
		t.Fatal(err)
	}
	if !e.Contains(errorResp.Err, "user or password invalid") {
		t.Fatal("invalid response")
	}
}
