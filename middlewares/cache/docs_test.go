// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package cache

import (
	"net/http"
	"testing"
	"time"
)

func TestText(t *testing.T) {
	mtime := time.Date(1979, 3, 1, 23, 42, 3, 14, time.Local)

	text := &Text{
		FileName:  "",
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
	text.ComposeFileName(r)

	if text.Name() != "/path/to/the/file.ext" {
		t.Fatal("wrong file name", text.Name())
	}

	if text.Sulfix() != ".ext" {
		t.Fatal("wrong extension")
	}

	text.Extension = ""
	text.ComposeFileName(r)

	if text.Name() != "/path/to/the/file.html" {
		t.Fatal("wrong file name", text.Name())
	}

	if text.Path() != "/1ad49718232128a1376ab2778b85dd49/path/to/the/file.html" {
		t.Fatal("wrong path name", text.Path())
	}

	if !text.Modification().Equal(mtime) {
		t.Fatal("wrong mtime")
	}

	if text.Sulfix() != "" {
		t.Fatal("wrong extension")
	}

	if text.RootDirHash() != "1ad49718232128a1376ab2778b85dd49" {
		t.Fatal("wrong hash")
	}

	newMtime := mtime.Add(time.Second)
	nt := text.New(newMtime)

	if nt.Name() != "/path/to/the/file.html" {
		t.Fatal("wrong file name", nt.Name())
	}

	if nt.Path() != "/1ad49718232128a1376ab2778b85dd49/path/to/the/file.html" {
		t.Fatal("wrong path name", nt.Path())
	}

	if !nt.Modification().Equal(newMtime) {
		t.Fatal("wrong mtime")
	}

	if nt.Sulfix() != "" {
		t.Fatal("wrong extension")
	}

	if nt.RootDirHash() != "1ad49718232128a1376ab2778b85dd49" {
		t.Fatal("wrong hash")
	}

	if !nt.OutDate(text) {
		t.Fatal("OutDate failed")
	}

	if text.OutDate(nt) {
		t.Fatal("OutDate failed")
	}
}

func TestImage(t *testing.T) {
	mtime := time.Date(1979, 3, 1, 23, 42, 3, 14, time.Local)

	img := &Image{
		FileName:  "",
		MTime:     mtime,
		Lang:      "pt-br",
		Mime:      "imagem/jpg",
		UserAgent: "fulero/1.0",
		Extension: ".jpg",
		Hash:      "aaa",
		Width:     2,
		Height:    2,
	}

	r, err := http.NewRequest("GET", "http://localhost/path/to/the/file", nil)
	if err != nil {
		t.Fatal(err)
	}
	img.ComposeFileName(r)

	if img.Name() != "/path/to/the/file.jpg" {
		t.Fatal("wrong file name", img.Name())
	}

	if img.Sulfix() != ".jpg" {
		t.Fatal("wrong extension")
	}

	img.Extension = ""
	img.ComposeFileName(r)

	if img.Name() != "/path/to/the/file.image" {
		t.Fatal("wrong file name", img.Name())
	}

	if img.Path() != "/ee0c7f03900375704a1cbd5b9b673c0d/path/to/the/file.image" {
		t.Fatal("wrong path name", img.Path())
	}

	if !img.Modification().Equal(mtime) {
		t.Fatal("wrong mtime")
	}

	if img.Sulfix() != "" {
		t.Fatal("wrong extension")
	}

	if img.RootDirHash() != "ee0c7f03900375704a1cbd5b9b673c0d" {
		t.Fatal("wrong hash")
	}

	newMtime := mtime.Add(time.Second)
	ni := img.New(newMtime)

	if ni.Name() != "/path/to/the/file.image" {
		t.Fatal("wrong file name", ni.Name())
	}

	if ni.Path() != "/ee0c7f03900375704a1cbd5b9b673c0d/path/to/the/file.image" {
		t.Fatal("wrong path name", ni.Path())
	}

	if !ni.Modification().Equal(newMtime) {
		t.Fatal("wrong mtime")
	}

	if ni.Sulfix() != "" {
		t.Fatal("wrong extension")
	}

	if ni.RootDirHash() != "ee0c7f03900375704a1cbd5b9b673c0d" {
		t.Fatal("wrong hash")
	}

	if !ni.OutDate(img) {
		t.Fatal("OutDate failed")
	}

	if img.OutDate(ni) {
		t.Fatal("OutDate failed")
	}
}
