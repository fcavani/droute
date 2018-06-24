// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/fcavani/e"
	log "github.com/fcavani/slog"

	"github.com/fcavani/droute/errhandler"
	"github.com/fcavani/droute/responsewriter"
	"gopkg.in/fcavani/httprouter.v2"
)

const ctxName string = "proxyredirdst"

// LoadBalance interface reuni the methods for a load balancing.
type LoadBalance interface {
	AddAddrs(method, path, ip string)
	Next(method, path string) string
	Remove(method, path, ip string)
}

// RedirDst redirects the resquest to a single address.
type RedirDst struct {
	dst string
}

// NewRedirDst creates a balancer with a single address.
func NewRedirDst(dst string) LoadBalance {
	return &RedirDst{
		dst: dst,
	}
}

// AddAddrs do nothing
func (rd *RedirDst) AddAddrs(method, path, ip string) {
	return
}

// Next return always the same address
func (rd *RedirDst) Next(method, path string) string {
	return rd.dst
}

// Remove do nothing
func (rd *RedirDst) Remove(method, path, ip string) {
	return
}

type ips struct {
	ips    []string
	actual int
}

// RoundRobin make a rounding robin list of ip addresses of destinies.
type RoundRobin struct {
	ips map[string]map[string]*ips
	lck sync.RWMutex
}

// NewRoundRobin creates a new rounding roubing list.
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{
		ips: make(map[string]map[string]*ips),
	}
}

// AddAddrs adds an ip to the list.
func (rr *RoundRobin) AddAddrs(method, path, ip string) {
	if path == "" {
		path = "/"
	}
	rr.lck.Lock()
	defer rr.lck.Unlock()
	m, ok := rr.ips[method]
	if !ok {
		m = make(map[string]*ips)
		rr.ips[method] = m
	}
	p, ok := m[path]
	if !ok {
		p = &ips{
			ips:    make([]string, 0, 10),
			actual: 0,
		}
		m[path] = p
	}
	for _, actual := range p.ips {
		if actual == ip {
			return
		}
	}
	p.ips = append(p.ips, ip)
}

// Next pops the next address from the list.
func (rr *RoundRobin) Next(method, path string) string {
	if path == "" {
		path = "/"
	}
	rr.lck.RLock()
	defer rr.lck.RUnlock()
	m, ok := rr.ips[method]
	if !ok {
		return ""
	}
	p := findPath(m, path)
	if p == nil {
		return ""
	}
	if len(p.ips) <= 0 {
		return ""
	}
	if p.actual >= len(p.ips) {
		p.actual = 0
	}
	defer func() {
		p.actual = p.actual + 1
	}()
	return p.ips[p.actual]
}

//Remove exclude a ip from the list of proxies
func (rr *RoundRobin) Remove(method, path, target string) {
	if path == "" {
		path = "/"
	}
	rr.lck.Lock()
	defer rr.lck.Unlock()
	m, ok := rr.ips[method]
	if !ok {
		return
	}
	p := findPath(m, path)
	if p == nil {
		return
	}
	for i, ip := range p.ips {
		if ip != target {
			continue
		}
		if i > 0 && i+1 < len(p.ips) {
			p.ips = append(p.ips[:i-1], p.ips[i+1:]...)
		} else if i == 0 && i+1 < len(p.ips) {
			p.ips = p.ips[i+1:]
		} else if i > 0 && i+1 == len(p.ips) {
			p.ips = p.ips[:i-1]
		} else if i == 0 && len(p.ips) == 1 {
			p.ips = p.ips[:0]
		}
		break
	}
}

// Balance is the handler that inserts in the context the next ip address.
func Balance(lb LoadBalance, handler responsewriter.HandlerFunc) responsewriter.HandlerFunc {
	return func(rw *responsewriter.ResponseWriter, req *http.Request) {
		lang := httprouter.ContentLang(req)
		path := req.URL.Path
		if lang != "" {
			path = strings.TrimPrefix(req.URL.Path, "/"+lang)
		}
		dst := lb.Next(req.Method, path)
		if dst == "" {
			log.Tag("router", "loadbalance").DebugLevel().Printf(
				"no proxy ip (%v, %v, %v)",
				req.Method,
				req.URL.Path,
				lang,
			)
			errhandler.ErrHandler(rw, http.StatusInternalServerError, e.New("no proxy ip address"))
			return
		}
		log.Tag("router", "loadbalance").DebugLevel().Printf(
			"Proxy ip: %v (%v, %v, %v)",
			dst,
			req.Method,
			req.URL.Path,
			lang,
		)
		req = req.WithContext(context.WithValue(req.Context(), ctxName, dst))
		handler(rw, req)
		code := rw.ResponseCode()
		// TODO: 500 is for server error not fatal...
		if code > 500 && code < 600 {
			log.Tag("router", "loadbalance").DebugLevel().Printf("remove proxy %v (%v)", dst, code)
			lb.Remove(req.Method, req.URL.Path, dst)
		}
	}
}

//TODO: random, least used...

func findPath(m map[string]*ips, path string) *ips {
	p, ok := m[path]
	if ok {
		return p
	}
	for k, p := range m {
		//if strings.Contains(k, "/*") || matchNamedParam(path, k) {
		if strings.Contains(k, path+"/*") || matchNamedParam(path, k) {
			return p
		}
	}
	return nil
}

func matchNamedParam(path, k string) bool {
	if path == "" || k == "" {
		return false
	}
	for i, p := range path {
		if i < len(k) && p != rune(k[i]) {
			if k[i] == ':' {
				return true
			}
			return false
		}
	}
	return path == k
}
