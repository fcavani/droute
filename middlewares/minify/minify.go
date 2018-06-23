// Copyright 2017 Felipe A. Cavani. All rights reserved.

// Package minify compact certains types of codes in a form readable only by
// the computer.
package minify

import (
	"net/http"

	"github.com/tdewolff/minify"
)

// Minify handler minifies certain types of code.
func Minify(m *minify.M, enable bool, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if !enable {
			f(w, r)
			return
		}

		resp := m.ResponseWriter(w, r)
		defer resp.Close()
		f(resp, r)
	}
}
