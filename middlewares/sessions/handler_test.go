// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package sessions

import (
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"github.com/fcavani/droute/responsewriter"
	"github.com/fcavani/droute/sessions"
	"github.com/fcavani/e"
	"github.com/gorilla/securecookie"
)

func TestHandler(t *testing.T) {

	sessionDB := sessions.NewMemSessions(60*time.Second, 20*time.Second)

	maxage := 60

	authKey, err := base64.StdEncoding.DecodeString("mqWkutRRHOw81Uwb4NVvuC/3Xqz04AzitOlsEZd5948=")
	if err != nil {
		t.Fatal(err)
	}
	encKey, err := base64.StdEncoding.DecodeString("M7zU2Euxi3mwtUrJtTzJ1S79F9DDC71sL+WNIdUYFyQ=")
	if err != nil {
		t.Fatal(err)
	}

	cookie := &Cookie{
		Cookie: &http.Cookie{
			Name:     "test",
			Path:     "/",
			Domain:   "localhost",
			Expires:  time.Now().Add(time.Duration(maxage) * time.Second).UTC(),
			MaxAge:   maxage,
			Secure:   true,
			HttpOnly: true,
		},
		SecureCookie: securecookie.New(authKey, encKey),
	}

	uuid := ""

	rw := responsewriter.NewResponseWriter()
	req, err := http.NewRequest("GET", "localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	h := Handler(sessionDB, cookie,
		func(w http.ResponseWriter, r *http.Request) {
			session := UserSession(r)
			if session.UUID() == "" {
				t.Fatal("uuid is empty")
			}
			t.Log(session.UUID())
			uuid = session.UUID()
			er := session.Set("a", 1)
			if er != nil {
				t.Fatal(er)
			}
			data, er := session.Get("a")
			if er != nil {
				t.Fatal(er)
			}
			if data.(int) != 1 {
				t.Fatal("data stored is wrong")
			}
			er = session.Del("a")
			if er != nil {
				t.Fatal(er)
			}
			er = session.Set("next", 1)
			if er != nil {
				t.Fatal(er)
			}
		},
	)
	h(rw, req)

	req, err = http.NewRequest("GET", "localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Cookie", rw.Header().Get("Set-Cookie"))
	rw = responsewriter.NewResponseWriter()

	h = Handler(sessionDB, cookie,
		func(w http.ResponseWriter, r *http.Request) {
			session := UserSession(r)
			if session.UUID() == "" {
				t.Fatal("uuid is empty")
			}
			t.Log(session.UUID())
			if uuid != session.UUID() {
				t.Fatal("uuid is not the same")
			}
			_, er := session.Get("a")
			if er != nil && !e.Equal(er, sessions.ErrSessionDataNotFound) {
				t.Fatal(er)
			} else if er == nil {
				t.Fatal("nil error")
			}
			data, er := session.Get("next")
			if er != nil {
				t.Fatal(er)
			}
			if data.(int) != 1 {
				t.Fatal("data stored is wrong")
			}
			er = session.Del("a")
			if er != nil && !e.Equal(er, sessions.ErrSessionDataNotFound) {
				t.Fatal(er)
			} else if er == nil {
				t.Fatal("nil error")
			}
		},
	)
	h(rw, req)
}
