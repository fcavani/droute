// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package cache

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/fcavani/afero"

	"github.com/fcavani/droute/responsewriter"
)

func TestStorage(t *testing.T) {
	start := time.Now()
	fs := afero.NewMemMapFs()
	stg := NewStorage("/", fs, time.Minute, time.Millisecond)

	mtime := time.Date(1979, 3, 1, 23, 42, 3, 14, time.Local)
	doc := &Text{
		MTime:     mtime,
		Lang:      "pt-br",
		Encoding:  "utf8",
		Mime:      "text/html",
		UserAgent: "fulero/1.0",
		Extension: ".ext",
		Hash:      "aaa",
	}
	r, err := http.NewRequest("GET", "http://localhost/path/to/the/file", nil)
	if err != nil {
		t.Fatal(err)
	}
	doc.ComposeFileName(r)

	header := make(map[string][]string)
	header["key01"] = []string{"value01"}
	header["key02"] = []string{"value02"}
	buf := bytes.NewBufferString("plain old data")

	err = stg.Store(doc, buf, header)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fs.Open(doc.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "plain old data" {
		t.Fatal("wrong data in file")
	}

	fh, err := fs.Open(doc.Path() + ".header")
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()

	hdr := make(map[string][]string)
	err = json.NewDecoder(fh).Decode(&hdr)
	if err != nil {
		t.Fatal(err)
	}

	v, found := hdr["key01"]
	if !found {
		t.Fatal("header fail")
	}
	if len(v) == 1 && v[0] != "value01" {
		t.Fatal("header fail")
	}
	v, found = hdr["key02"]
	if !found {
		t.Fatal("header fail")
	}
	if len(v) == 1 && v[0] != "value02" {
		t.Fatal("header fail")
	}

	resp := responsewriter.NewResponseWriter()
	err = stg.ReadTo(doc, resp, 200)
	if err != nil {
		t.Fatal(err)
	}

	data = resp.Bytes()
	if string(data) != "plain old data" {
		t.Fatal("wrong data in file")
	}

	hdr = resp.Header()
	v, found = hdr["Key01"]
	if !found {
		t.Fatal("header fail")
	}
	if len(v) == 1 && v[0] != "value01" {
		t.Fatal("header fail")
	}
	v, found = hdr["Key02"]
	if !found {
		t.Fatal("header fail")
	}
	if len(v) == 1 && v[0] != "value02" {
		t.Fatal("header fail")
	}

	newDoc, err := stg.GetMTime(doc)
	if err != nil {
		t.Fatal(err)
	}

	fi, err := fs.Stat(doc.Path())
	if err != nil {
		t.Fatal(err)
	}

	if !newDoc.Modification().After(start) || !fi.ModTime().Equal(newDoc.Modification()) {
		t.Fatal("wrong mtime", newDoc.Modification(), mtime)
	}

	stg.Close()
}
