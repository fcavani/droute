// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package scheme

import (
	"net/http"
	"net/url"

	eh "github.com/fcavani/droute/errhandler"
	h "github.com/fcavani/http"
)

type RedirType uint8

const (
	ToHttps RedirType = iota
	ToHttp
	DontCare
)

// Redirect change the scheme and redirect the request. If the scheme do not
// change call the next handler.
func Redirect(rt RedirType, root, port string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if root == "/" {
			root = ""
		}

		u, err := h.Url(r, root)
		if err != nil {
			eh.ErrHandler(w, http.StatusInternalServerError, err)
			return
		}

		var newu *url.URL
		switch rt {
		case ToHttps:
			newu, err = h.UrlSecure(u, port)
			if err != nil {
				eh.ErrHandler(w, http.StatusInternalServerError, err)
				return
			}
		case ToHttp:
			newu, err = h.UrlUnsecure(u, port)
			if err != nil {
				eh.ErrHandler(w, http.StatusInternalServerError, err)
				return
			}
		}

		if newu.Scheme != u.Scheme {
			http.Redirect(w, r, newu.String(), 303)
			return
		}

		handler(w, r)
	}
}

// ChangeProto modify the url scheme.
func ChangeProto(rt RedirType, root, port string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if root == "/" {
			root = ""
		}

		u, err := h.Url(r, root)
		if err != nil {
			eh.ErrHandler(w, http.StatusInternalServerError, err)
			return
		}

		var newu *url.URL
		switch rt {
		case ToHttps:
			newu, err = h.UrlSecure(u, port)
			if err != nil {
				eh.ErrHandler(w, http.StatusInternalServerError, err)
				return
			}
		case ToHttp:
			newu, err = h.UrlUnsecure(u, port)
			if err != nil {
				eh.ErrHandler(w, http.StatusInternalServerError, err)
				return
			}
		}

		if newu.Scheme != u.Scheme {
			if r.URL.Scheme == "https" && newu.Scheme == "http" {
				r.TLS = nil
			}
			r.URL = newu
		}

		handler(w, r)
	}
}
