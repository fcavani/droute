// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"net/http"
	"time"

	"github.com/fcavani/afero"
	"gopkg.in/fcavani/httprouter.v2"

	"github.com/fcavani/droute/errhandler"
	"github.com/fcavani/droute/middlewares/cache"
	"github.com/fcavani/droute/middlewares/compress"
	"github.com/fcavani/droute/middlewares/expire"
	"github.com/fcavani/droute/middlewares/request"
	h "github.com/fcavani/http"
)

// NewStaticRouter is a fileserver for a static dir.
func NewStaticRouter(prefix, dir string, fs afero.Fs, expTime time.Duration, rc *request.Config, formsizefile int64, cs *cache.Storage) (router *httprouter.Router) {
	router = httprouter.New()
	router.GET("/*filename",
		request.Handler(rc, formsizefile,
			cache.Cache(expTime, cs, cache.MetaFS(fs, prefix),
				compress.Compress(
					expire.Expire(expTime,
						Adaptor(
							http.FileServer(
								http.Dir(dir),
							),
						),
					),
				),
			),
		),
	)
	return
}

//NewRedirHostRouter is a router that redirect to another host the request.
func NewRedirHostRouter(dst string) (r *httprouter.Router) {
	r = httprouter.New()
	r.GET("/*filename",
		RedirHost("", dst),
	)
	return
}

// RedirHost redirect from one host to another without modify the other parameters
// of the url.
func RedirHost(root, dst string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := h.Url(r, root)
		if err != nil {
			errhandler.ErrHandler(w, http.StatusInternalServerError, err)
			return
		}
		u.Host = dst
		http.Redirect(w, r, u.String(), 303)
	}
}

// Adaptor is used for plug a http.Handler with a http.HandlerFunc.
func Adaptor(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		return
	}
}
