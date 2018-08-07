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

package router

import (
	"net/http"
	"time"

	"github.com/fcavani/droute/errhandler"
	"github.com/fcavani/droute/list"
	"github.com/fcavani/droute/responsewriter"
	"github.com/fcavani/e"
	h "github.com/fcavani/http"
	log "github.com/fcavani/slog"
)

func IPBlock(deny list.List, handler responsewriter.HandlerFunc) responsewriter.HandlerFunc {
	return func(w *responsewriter.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			log.DebugLevel().Tag("ipblock").Println("ip block took:", time.Since(start))
		}()
		ip, _ := h.RemoteIP(r)
		if ip != "" {
			if deny != nil {
				if !deny.Exist(ip) {
					log.DebugLevel().Tag("ipblock").Printf("Allowed: %v", ip)
					handler(w, r)
					return
				}
			}
		}
		log.DebugLevel().Tag("ipblock").Printf("Denied (default): %v", ip)
		errhandler.ErrHandler(w, http.StatusForbidden, e.New("ip deny"))
		return
	}
}
