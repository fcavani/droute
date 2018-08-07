// Copyright © 2017 Felipe A. Cavani <fcavani@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package iplists control the access by the referrer address or ip address.
package iplists

import (
	"net/http"
	"time"

	"github.com/fcavani/droute/errhandler"
	"github.com/fcavani/droute/list"
	"github.com/fcavani/droute/middlewares/request"
	"github.com/fcavani/e"
	"github.com/fcavani/net/dns"
	log "github.com/fcavani/slog"
)

type DefaultRule uint8

const (
	DENY DefaultRule = iota
	ALLOW
)

// Referrer checks the referrer url
func Referrer(def DefaultRule, allowed, deny list.List, fbc *FBCrawler, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			log.DebugLevel().Tag("referrer").Println("Referrer took:", time.Since(start))
		}()
		rh := request.Request(r)
		ref := rh.Referrer()
		if ref != nil {
			if allowed != nil {
				if allowed.Exist(ref.String()) {
					log.DebugLevel().Tag("referrer").Printf("Allowed: %v", ref)
					handler(w, r)
					return
				}
			}
			if deny != nil {
				if deny.Exist(ref.String()) {
					if fbc != nil && isFB(fbc, ref.String()) {
						log.DebugLevel().Tag("referrer").Printf("Allowed (is facebook): %v", ref)
						handler(w, r)
						return
					}
					log.DebugLevel().Tag("referrer").Printf("Denied: %v", ref)
					errhandler.ErrHandler(w, http.StatusForbidden, e.New("referrer deny"))
					return
				}
			}
		}
		if def == ALLOW {
			log.DebugLevel().Tag("referrer").Printf("Allowed (default): %v", ref)
			handler(w, r)
			return
		}
		if ip := rh.IP(); ip != "" && fbc != nil && fbc.HaveIP(ip) {
			log.DebugLevel().Tag("referrer").Printf("Allowed (is facebook): %v", ref)
			handler(w, r)
			return
		}
		log.DebugLevel().Tag("referrer").Printf("Denied (default): %v", ref)
		errhandler.ErrHandler(w, http.StatusForbidden, e.New("referrer deny"))
		return
	}
}

func isFB(fbc *FBCrawler, ref string) bool {
	if ip, err := dns.Resolve(ref); err == nil {
		return fbc.HaveIP(ip)
	}
	return false
}
