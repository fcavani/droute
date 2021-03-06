// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package redirect

import "net/http"

// Redirect redirects from one url to another.
func Redirect(code int, url string) http.HandlerFunc {
	if code <= 0 {
		code = 303
	}
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, code)
	}
}
