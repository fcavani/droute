// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"net/http"
	"strings"

	"github.com/fcavani/e"
	utilNet "github.com/fcavani/net"
)

//ErrInvIPAddr is an error
const ErrInvIPAddr = "invalid ip address"

// RemoteIP return the ip addrs of the caller.
func RemoteIP(r *http.Request) (string, error) {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		host, _, err := utilNet.SplitHostPort(r.RemoteAddr)
		if err != nil && !e.Equal(err, utilNet.ErrCantFindPort) {
			return "", e.Forward(err)
		}
		return host, nil
	}
	if hdrForwardedFor != "" {
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
			if !utilNet.IsValidIpv4(parts[i]) && !utilNet.IsValidIpv6(parts[i]) {
				return "", e.New(ErrInvIPAddr)
			}
		}
		// TODO: should return first non-local address
		return parts[0], nil
	}
	if !utilNet.IsValidIpv4(hdrRealIP) && !utilNet.IsValidIpv6(hdrRealIP) {
		return "", e.New(ErrInvIPAddr)
	}
	return hdrRealIP, nil
}
