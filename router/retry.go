// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"net/http"

	"github.com/fcavani/droute/responsewriter"
)

// Retry try multiple time to get a correct response from the handler.
func Retry(times int, handler responsewriter.HandlerFunc) responsewriter.HandlerFunc {
	return func(rw *responsewriter.ResponseWriter, req *http.Request) {
		for i := 0; i < times; i++ {
			handler(rw, req)
			code := rw.ResponseCode()
			if !(code >= 500 && code < 600) {
				break
			}
		}
	}
}
