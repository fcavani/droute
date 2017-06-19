// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"net/http"

	"github.com/fcavani/log"
	"golang.org/x/net/idna"
)

//HostSwitch selects the router associated with one hostname.
type HostSwitch map[string]http.Handler

// ServeHTTP satisfy the interface for http package.
func (hs HostSwitch) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host, err := idna.ToUnicode(r.Host)
	if err != nil {
		log.Tag("hostswitch", "services").Errorf("Error in host name: %v", err)
		http.Error(w, "Invalid host name", 403)
		return
	}
	if handler := hs[host]; handler != nil {
		handler.ServeHTTP(w, r)
	} else {
		http.Error(w, "Forbidden", 403) // Or Redirect
	}
}

// Set store a new router for that domain or overwrite the router for the domain.
func (hs HostSwitch) Set(domain string, router http.Handler) {
	hs[domain] = router
}
