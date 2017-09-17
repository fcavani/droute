// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

// Package expire contains the expire midleware that adds to the http header
// the expire field.
package hsts

import (
	"net/http"
)

// Expire implements the expire midleware to put the expire field in the http
// header.
func HSTS(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// https://www.chromium.org/hsts
		// https://hstspreload.org/
		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		f(w, r)
	}
}
