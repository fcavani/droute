// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

// Package expire contains the expire midleware that adds to the http header
// the expire field.
package expire

import (
	"net/http"
	"time"
)

// Expire implements the expire midleware to put the expire field in the http
// header.
func Expire(expire time.Duration, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if expire != 0 {
			w.Header().Add("Expires", time.Now().Add(expire).UTC().Format(time.RFC1123))
		}
		f(w, r)
	}
}
