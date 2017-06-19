// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package cache

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fcavani/afero"
	"github.com/fcavani/e"

	"github.com/fcavani/droute/responsewriter"
)

// Criar um meta que retorna o mtime adequado

func newDoc(mtime time.Time, r *http.Request) Document {
	txt := &Text{
		FileName:  "",
		MTime:     mtime,
		Lang:      "pt-br",
		Encoding:  "utf8",
		Mime:      "text/plain",
		UserAgent: "fulero/1.0",
		Extension: ".txt",
		Hash:      "aaa",
	}
	txt.ComposeFileName(r)
	return txt
}

func makeMeta(mtime time.Time, er error) Meta {
	return func(r *http.Request) (Document, *http.Request, error) {
		if er != nil {
			return nil, nil, er
		}
		return newDoc(mtime, r), r, nil
	}
}

func makeCached(fs afero.Fs, filename string, mtime time.Time, r *http.Request) error {
	doc := newDoc(mtime, r)
	filename = filepath.Join("/", doc.RootDirHash(), filename)
	header := filename + ".header"

	fs.RemoveAll(filename)
	fs.RemoveAll(header)

	f, err := fs.Create(filename)
	if err != nil {
		return e.Forward(err)
	}
	_, err = f.WriteString("42")
	if err != nil {
		f.Close()
		fs.Remove(filename)
		return e.Forward(err)
	}
	f.Close()
	err = fs.Chtimes(filename, time.Now(), mtime)
	if err != nil {
		fs.Remove(filename)
		return e.Forward(err)
	}
	// Header
	fh, err := fs.Create(header)
	if err != nil {
		return e.Forward(err)
	}
	defer fh.Close()

	h := make(map[string][]string)

	err = json.NewEncoder(fh).Encode(h)
	if err != nil {
		return e.Forward(err)
	}

	err = fh.Sync()
	if err != nil {
		return e.Forward(err)
	}

	return nil
}

var fs afero.Fs
var stg *Storage

func TestStoreStart(t *testing.T) {
	fs = afero.NewMemMapFs()
	stg = NewStorage("", fs, time.Minute, time.Millisecond)
}

const filename = "file.txt"

// Não existe cache...
func TestCacheNotFound(t *testing.T) {
	missed := false
	// Actual date of the file
	mtime := time.Date(1979, 3, 1, 23, 42, 3, 14, time.Local)

	r, err := http.NewRequest("GET", "http://localhost/"+filename, nil)
	if err != nil {
		t.Fatal(err)
	}

	f := Cache(
		time.Minute,
		stg,
		makeMeta(mtime, nil),
		func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(200)
			fmt.Fprintln(rw, "Oi!")
			missed = true
		},
	)
	rw := responsewriter.NewResponseWriter()
	f(rw, r)
	if !missed {
		t.Fatal("Cache hit!")
	}
}

func TestCacheMiss(t *testing.T) {
	missed := false
	// Actual date of the file
	mtime := time.Date(1979, 3, 1, 23, 42, 3, 14, time.Local)
	// Date of the cache
	past := time.Date(1979, 3, 1, 23, 42, 3, 13, time.Local)
	// mtime > past => cache miss
	r, err := http.NewRequest("GET", "http://localhost/"+filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = makeCached(fs, filename, past, r)
	if err != nil {
		t.Fatal(err)
	}
	f := Cache(
		time.Minute,
		stg,
		makeMeta(mtime, nil),
		func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(200)
			fmt.Fprintln(rw, "Oi!")
			missed = true
		},
	)
	// afero.Walk(fs, "/", func(path string, info os.FileInfo, err error) error {
	// 	fmt.Println(path, info)
	// 	return nil
	// })
	rw := responsewriter.NewResponseWriter()
	f(rw, r)
	if !missed {
		t.Fatal("Cache hit!")
	}
}

// Cache existe

func TestCacheHit(t *testing.T) {
	missed := false
	// Actual date of the file
	mtime := time.Date(1979, 3, 1, 23, 42, 3, 14, time.Local)
	// Date of the cache
	past := time.Date(1979, 3, 1, 23, 42, 3, 15, time.Local)
	// mtime > past => cache miss
	r, err := http.NewRequest("GET", "http://localhost/"+filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = makeCached(fs, filename, past, r)
	if err != nil {
		t.Fatal(err)
	}
	f := Cache(
		time.Minute,
		stg,
		makeMeta(mtime, nil),
		func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(200)
			fmt.Fprintln(rw, "Oi!")
			missed = true
		},
	)
	rw := responsewriter.NewResponseWriter()
	f(rw, r)
	if missed {
		t.Fatal("Cache miss!")
	}
}

// // Arquivo original foi modificado
//
// func TestFileChanged(t *testing.T) {
//
// }

// If-Modified-Since

func TestIfModifiedSince(t *testing.T) {
	missed := false
	// Actual date of the file
	mtime := time.Date(1979, 3, 1, 23, 42, 3, 14, time.UTC)
	// Date of the cache
	past := time.Date(1979, 3, 1, 23, 43, 3, 15, time.UTC)
	// mtime > past => cache miss
	r, err := http.NewRequest("GET", "http://localhost/"+filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	f := Cache(
		time.Minute,
		stg,
		makeMeta(mtime, nil),
		func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(200)
			fmt.Fprintln(rw, "Oi!")
			missed = true
		},
	)
	rw := responsewriter.NewResponseWriter()
	r.Header.Add("If-Modified-Since", past.Format(timeFormat))
	f(rw, r)
	if missed {
		t.Fatal("Cache miss!")
	}
	if rw.Header().Get("X-Cache") != "since" {
		t.Fatal("not cached")
	}
	if rw.ResponseCode() != http.StatusNotModified {
		t.Fatal("wrong response code", rw.ResponseCode())
	}
}

func TestIfModifiedSince2(t *testing.T) {
	missed := false
	// Actual date of the file
	mtime := time.Date(1979, 3, 1, 23, 42, 3, 14, time.UTC)
	// Date of the cache
	past := time.Date(1979, 3, 1, 23, 41, 3, 13, time.UTC)
	// mtime > past => cache miss
	r, err := http.NewRequest("GET", "http://localhost/"+filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	f := Cache(
		time.Minute,
		stg,
		makeMeta(mtime, nil),
		func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(200)
			fmt.Fprintln(rw, "Oi!")
			missed = true
		},
	)
	rw := responsewriter.NewResponseWriter()
	r.Header.Add("If-Modified-Since", past.Format(timeFormat))
	f(rw, r)
	if rw.Header().Get("X-Cache") != "true" {
		t.Fatal("not cached")
	}
	if rw.ResponseCode() != http.StatusOK {
		t.Fatal("wrong response code", rw.ResponseCode())
	}
}

// f response 200 series e 500 series
// Response f 0

func TestResponseZero(t *testing.T) {
	// Actual date of the file
	mtime := time.Date(1979, 3, 1, 23, 42, 3, 14, time.Local)
	// Date of the cache
	past := time.Date(1979, 3, 1, 23, 42, 3, 13, time.Local)
	// mtime > past => cache miss
	r, err := http.NewRequest("GET", "http://localhost/"+filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = makeCached(fs, filename, past, r)
	if err != nil {
		t.Fatal(err)
	}
	f := Cache(
		time.Minute,
		stg,
		makeMeta(mtime, nil),
		func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(rw, "Oi!")
			rw.WriteHeader(0)
		},
	)
	rw := responsewriter.NewResponseWriter()
	f(rw, r)
	if rw.ResponseCode() != 200 {
		t.Fatal("wrong response code")
	}
}

func TestResponse500(t *testing.T) {
	// Actual date of the file
	mtime := time.Date(1979, 3, 1, 23, 42, 3, 14, time.Local)
	// Date of the cache
	past := time.Date(1979, 3, 1, 23, 42, 3, 13, time.Local)
	// mtime > past => cache miss
	r, err := http.NewRequest("GET", "http://localhost/"+filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = makeCached(fs, filename, past, r)
	if err != nil {
		t.Fatal(err)
	}
	f := Cache(
		time.Minute,
		stg,
		makeMeta(mtime, nil),
		func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(rw, "Oi!")
			rw.WriteHeader(501)
		},
	)
	rw := responsewriter.NewResponseWriter()
	f(rw, r)
	if rw.ResponseCode() != 501 {
		t.Fatal("wrong response code", rw.ResponseCode())
	}
}

func TestError(t *testing.T) {
	// Actual date of the file
	mtime := time.Date(1979, 3, 1, 23, 42, 3, 14, time.Local)
	// Date of the cache
	past := time.Date(1979, 3, 1, 23, 42, 3, 13, time.Local)
	// mtime > past => cache miss
	r, err := http.NewRequest("GET", "http://localhost/"+filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = makeCached(fs, filename, past, r)
	if err != nil {
		t.Fatal(err)
	}
	f := Cache(
		time.Minute,
		stg,
		makeMeta(mtime, nil),
		func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(0)
		},
	)
	rw := responsewriter.NewResponseWriter()
	f(rw, r)
	if rw.ResponseCode() != 500 {
		t.Fatal("wrong response code", rw.ResponseCode())
	}
	var merr msgErr
	err = json.NewDecoder(rw).Decode(&merr)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(merr.Err, "response body lenght is zero") {
		t.Fatal("wrong error response")
	}
}
