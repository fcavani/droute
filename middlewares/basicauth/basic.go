// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package basicauth

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/fcavani/e"
)

const ErrNoAuthHeader = "no authorization header found"
const ErrInvalidAuthHeader = "invalid authorization header"
const ErrInvalidAuthString = "invalid authorization string"
const ErrInvalidAuthScheme = "invalid authorization scheme"
const ErrDecodeUsrPass = "failed to decode the user and password string"
const ErrInvalidUsrPassString = "invalid user pass string"

// Basic gets the first scheme in Authorization header and return
// the userid and the password.
func Basic(r *http.Request) (user, pass string, err error) {
	auth, ok := r.Header["Authorization"]
	if !ok {
		return "", "", e.New(ErrNoAuthHeader)
	}
	user, pass, err = basic(auth)
	if err != nil {
		err = e.Forward(err)
		return
	}
	return
}

func basic(auth []string) (user, pass string, err error) {
	if len(auth) <= 0 {
		return "", "", e.New(ErrInvalidAuthHeader)
	}
	s := strings.Fields(auth[0])
	if len(s) != 2 {
		return "", "", e.New(ErrInvalidAuthString)
	}
	if s[0] != "Basic" {
		return "", "", e.New(ErrInvalidAuthScheme)
	}
	data, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return "", "", e.Push(e.New(err), ErrDecodeUsrPass)
	}
	usrpass := strings.Split(string(data), ":")
	if len(usrpass) != 2 {
		return "", "", e.New(ErrInvalidUsrPassString)
	}
	return usrpass[0], usrpass[1], nil
}
