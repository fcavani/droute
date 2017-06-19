// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"net/http"

	"github.com/fcavani/droute/errhandler"
	"github.com/fcavani/droute/responsewriter"

	"github.com/fcavani/e"
	"github.com/fcavani/log"
	"github.com/sony/gobreaker"
)

// CircuitBrake brakes the connection from the proxy to the server if the server
// dies.
func CircuitBrake(cbs map[string]*gobreaker.CircuitBreaker, handler responsewriter.HandlerFunc) responsewriter.HandlerFunc {
	return func(rw *responsewriter.ResponseWriter, req *http.Request) {
		proxy := req.Context().Value("proxyredirdst").(string)
		cb, found := cbs[proxy]
		if !found {
			st := gobreaker.Settings{
				Name: proxy,
			}
			cb = gobreaker.NewCircuitBreaker(st)
			cbs[proxy] = cb
		}
		_, err := cb.Execute(func() (interface{}, error) {
			handler(rw, req)
			code := rw.ResponseCode()
			if code >= 500 && code < 600 {
				return nil, e.New("server fail")
			}
			return nil, nil
		})
		if err != nil {
			delete(cbs, proxy)
			errhandler.ErrHandler(rw, http.StatusInternalServerError, e.Forward(err))
			log.DebugLevel().Tag("router", "circuitbrake").Println(err)
			return
		}
	}
}
