// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package sessions

import (
	"context"
	"net/http"

	"github.com/fcavani/droute/errhandler"
	"github.com/fcavani/droute/sessions"
	"github.com/fcavani/e"
	log "github.com/fcavani/slog"
)

func UserSession(r *http.Request) sessions.Session {
	return r.Context().Value("session").(sessions.Session)
}

func Handler(s sessions.Sessions, c *Cookie, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid, err := c.UUID(r)
		if err != nil {
			uuid, err = c.New(w)
			if err != nil {
				errhandler.ErrHandler(w, 500, e.Push(err, "can't initiate the session"))
				return
			}
		}
		sess, err := s.Restore(uuid)
		if err != nil && e.Equal(err, sessions.ErrSessNotFound) {
			sess, err = s.New(uuid)
			if err != nil {
				errhandler.ErrHandler(w, 500, e.Push(err, "can't create a new session"))
				return
			}
		} else if err != nil {
			errhandler.ErrHandler(w, 500, e.Push(err, "can't start the session"))
			return
		}
		defer func() {
			er := s.Return(sess)
			if er != nil {
				//errhandler.ErrHandler(w, 500, e.Push(err, "can't return the session"))
				log.Tag("session", "handler", "return").Errorln(er)
				return
			}
		}()
		r = r.WithContext(context.WithValue(r.Context(), "session", sess))
		handler(w, r)
	}
}
