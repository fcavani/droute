// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

// Package cache is a midware thats stores one signature, composed of
// fields with information about the document that is accessed,
// with that signature one document is identified and the validit of
// it is considered for the sacke of cache.
package cache

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/fcavani/e"
	"github.com/fcavani/log"

	"github.com/fcavani/droute/responsewriter"
)

// Meta function collects all data about the document and put it a struct
// suitable to the cache handler.
type Meta func(r *http.Request) (Document, *http.Request, error)

const timeFormat = "Mon, 02 Jan 2006 15:04:05 MST"

// Cache handler check with version do to the user, cached or a new one.
// Cache sets too the header cache specific values.
func Cache(expire time.Duration, cs *Storage, meta Meta, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Reunir caracteristicas
		// original
		m, r, err := meta(r)
		if err != nil {
			errHandler(w, http.StatusInternalServerError, err)
			return
		}
		// Check cache metadata
		cached, err := cs.GetMTime(m)
		if err != nil {
			log.Tag("httpcache").InfoLevel().Printf("Cache miss: %v, %v", m.Path(), e.Trace(err))
			errHandler(w, http.StatusInternalServerError, e.Forward(exec(expire, cs, m, w, r, f)))
			return
		}
		// Test the header information. If header is newer than mtime...
		if t, er := time.Parse(timeFormat, r.Header.Get("If-Modified-Since")); er == nil {
			if m.Modification().Before(t.Add(1 * time.Second)) {
				log.Tag("httpcache").InfoLevel().Printf("Cache not modified: %v", m.Path())
				//Cache control
				w.Header().Add("X-Cache", "since")
				cacheCtrl(w, expire, m, http.StatusNotModified)
				return
			}
		}
		// Avaliar caracteristicas
		if m.OutDate(cached) {
			log.Tag("httpcache").InfoLevel().Printf("Cache miss: %v, outdated", m.Path())
			errHandler(w, http.StatusInternalServerError, e.Forward(exec(expire, cs, m, w, r, f)))
			return
		}
		//Cache control
		w.Header().Add("X-Cache", "true")
		cacheCtrl(w, expire, m, 200)
		// Read cache
		err = cs.ReadTo(m, w)
		if err != nil {
			log.Tag("httpcache").InfoLevel().Printf("Cache miss: %v, %v", m.Path(), e.Trace(err))
			errHandler(w, http.StatusInternalServerError, e.Forward(exec(expire, cs, m, w, r, f)))
			return
		}
		log.Tag("httpcache").InfoLevel().Printf("Cache hit: %v", m.Path())
	}
}

func exec(expire time.Duration, cs *Storage, m Document, w http.ResponseWriter, r *http.Request, f http.HandlerFunc) error {
	// TODO: Escrever um responsewriter que já manda para w e para o cache ao
	// mesmo tempo, tentar diminuir o número de buffers.
	var err error
	resp := responsewriter.NewResponseWriter()
	f(resp, r)
	l := resp.Len()
	if l == 0 {
		return e.New("response body lenght is zero")
	}
	buffer := resp.Bytes()
	cache := make([]byte, len(buffer))
	copy(cache, buffer)
	tow := bytes.NewBuffer(cache)
	resp.Header().Del("Content-Length")
	resp.Header().Add("Content-Length", strconv.FormatInt(int64(l), 10))
	code := resp.ResponseCode()
	if code == 0 || (code >= 200 && code < 300) {
		err = cs.Store(m, resp, resp.Header())
		if err != nil {
			return e.Forward(err)
		}
	}
	for k, v := range resp.Header() {
		for _, item := range v {
			w.Header().Add(k, item)
		}
	}
	if code == 0 {
		code = 200
	}
	//Cache control
	w.Header().Add("X-Cache", "false")
	cacheCtrl(w, expire, m, code)
	_, err = io.Copy(w, tow)
	if err != nil {
		return e.Forward(err)
	}
	log.Tag("httpcache").InfoLevel().Printf("Cache miss: %v", m.Path())
	return nil
}

func cacheCtrl(w http.ResponseWriter, expire time.Duration, m Document, code int) {
	if code != 0 && !(code >= 200 && code < 300) {
		w.WriteHeader(code)
		return
	}
	if expire != 0 {
		maxAge := strconv.FormatFloat(expire.Seconds(), 'f', 0, 64)
		w.Header().Add("Cache-Control", "max-age="+maxAge+", public")
		w.Header().Add("Expires", time.Now().Add(expire).UTC().Format(http.TimeFormat))
	}
	w.Header().Add("Last-Modified", m.Modification().UTC().Format(http.TimeFormat))
	w.Header().Add("Etag", m.RootDirHash())
	if code == 0 {
		return
	}
	w.WriteHeader(code)
}

type msgErr struct {
	Err string
}

func errHandler(w http.ResponseWriter, code int, err error) {
	if err == nil {
		return
	}
	log.DebugLevel().Tag("httpcache", "error").Println(err)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	resp := msgErr{
		Err: err.Error(),
	}
	er := json.NewEncoder(w).Encode(resp)
	if er != nil {
		log.Tag("router", "server", "proxy").Error(er)
	}
}
