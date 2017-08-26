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
	"github.com/fcavani/e"
	"github.com/gorilla/securecookie"
)

func TestCookieNew(t *testing.T) {
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

	rw := responsewriter.NewResponseWriter()

	uuid, err := cookie.New(rw)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("GET", "localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Cookie", rw.Header().Get("Set-Cookie"))

	rw = responsewriter.NewResponseWriter()

	uuidRetrived, err := cookie.UUID(req)
	if err != nil {
		t.Fatal(err)
	}

	if uuid != uuidRetrived {
		t.Fatal("wrong uuid", uuid, uuidRetrived)
	}

	req, err = http.NewRequest("GET", "localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cookie.UUID(req)
	if err != nil && !e.Contains(err, "not present") {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("nil error")
	}

}
