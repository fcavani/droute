// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package sessions

import (
	"net/http"
	"time"

	"github.com/fcavani/e"
	"github.com/fcavani/rand"
	"github.com/gorilla/securecookie"
)

type Cookie struct {
	*http.Cookie
	*securecookie.SecureCookie
}

func (c Cookie) UUID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(c.Cookie.Name)
	if err != nil {
		return "", e.Forward(err)
	}
	// Get the cookie
	values := make(map[string]string)
	err = c.Decode(c.Cookie.Name, cookie.Value, &values)
	if err != nil {
		return "", e.Forward(err)
	}
	return values["uuid"], nil
}

func (c Cookie) New(w http.ResponseWriter) (string, error) {
	uuid, err := rand.Uuid()
	if err != nil {
		return "", e.Forward(err)
	}
	values := map[string]string{
		"uuid": uuid,
	}
	val, err := c.Encode(c.Cookie.Name, values)
	if err != nil {
		return "", e.Forward(err)
	}
	cookie := &http.Cookie{
		Name:     c.Cookie.Name,
		Value:    val,
		Path:     c.Cookie.Path,
		Domain:   c.Cookie.Domain,
		Expires:  time.Now().Add(time.Duration(c.Cookie.MaxAge) * time.Second).UTC(),
		MaxAge:   c.Cookie.MaxAge,
		Secure:   c.Cookie.Secure,
		HttpOnly: c.Cookie.HttpOnly,
	}
	http.SetCookie(w, cookie)
	return uuid, nil
}
