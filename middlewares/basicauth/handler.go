// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package basicauth

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"hash"
	"net/http"
	"projects/http/auth"
	"strings"

	"github.com/fcavani/droute/errhandler"
	"github.com/fcavani/e"
)

type Validator interface {
	Validate(user, pass string) bool
}

var Hash map[string]hash.Hash

func init() {
	Hash = make(map[string]hash.Hash)
	Hash["plain"] = nil
	Hash["md5"] = md5.New()
	Hash["sha1"] = sha1.New()
	Hash["sha224"] = sha256.New224()
	Hash["sha256"] = sha256.New()
	Hash["sha384"] = sha512.New384()
	Hash["sha512"] = sha512.New()
}

type Users map[string]string

func splitHash(p string) (ht, pass string) {
	//ht: hass type
	i := strings.Index(p, ":")
	if i == -1 {
		return
	}
	if i+1 >= len(p) {
		return
	}
	ht = p[:i]
	pass = p[i+1:]
	return
}

const (
	// ErrInvName is the error for invalid name.
	ErrInvName = "invalid user name"
	// ErrBlankPass is the error for blank password
	ErrBlankPass = "blank password isn't allowed"
	//ErrInvHash is the error for invalid hash type
	ErrInvHash = "invalid password hash type"
)

func (u Users) SetPass(user, plainpass, hashtype string) error {
	if user == "" {
		return e.New(ErrInvName)
	}
	if plainpass == "" {
		return e.New(ErrBlankPass)
	}
	pass := "PLAIN:" + plainpass
	h, found := Hash[strings.ToLower(hashtype)]
	if !found {
		return e.New(ErrInvHash)
	}
	if h != nil {
		pass = strings.ToTitle(hashtype) + ":" + base64.StdEncoding.EncodeToString(h.Sum([]byte(plainpass)))
	}
	u[user] = pass
	return nil
}

func (u Users) SetRaw(user, pass string) error {
	if user == "" {
		return e.New(ErrInvName)
	}
	if pass == "" {
		return e.New(ErrBlankPass)
	}
	u[user] = pass
	return nil
}

func (u Users) Validate(user, plainpass string) bool {
	p, ok := u[user]
	if !ok {
		return false
	}
	ht, storedpass := splitHash(p)
	if ht == "" {
		return false
	}
	h := Hash[strings.ToLower(ht)]
	if h == nil {
		// Plain text password
		return storedpass == plainpass
	}
	sum := base64.StdEncoding.EncodeToString(h.Sum([]byte(plainpass)))
	return sum == storedpass
}

func BasicAuth(realm string, v Validator, handler http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		user, pass, err := Basic(r)
		if e.Equal(err, auth.ErrNoAuthHeader) {
			rw.Header().Add("WWW-Authenticate", "Basic realm=\""+realm+"\"")
			rw.WriteHeader(401)
			return
		} else if err != nil {
			errhandler.ErrHandler(rw, http.StatusForbidden, err)
			return
		}
		if !v.Validate(user, pass) {
			errhandler.ErrHandler(rw, http.StatusForbidden, e.New("user or password invalid"))
			return
		}
		handler(rw, r)
	}
}
