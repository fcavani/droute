// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package cache

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Document interface give access to cache control functions for the documents
// struct.
type Document interface {
	OutDate(cached Document) bool
	Name() string
	Path() string
	Modification() time.Time
	RootDirHash() string
	New(m time.Time) Document
	Sulfix() string
}

// Text represents the document. The local version.
type Text struct {
	FileName  string
	MTime     time.Time
	Lang      string
	Encoding  string
	Mime      string
	UserAgent string
	// Extension must start with a dot
	Extension string
	//ResponseCode string
	// Hash the template data.
	Hash string
}

// ComposeFileName creates the file name from the data find in the request.
func (t *Text) ComposeFileName(r *http.Request) {
	t.FileName = filepath.Join("/", dir(r.URL), t.genName(name(r.URL)))
}

// Path return the full path of the cache
func (t *Text) Path() string {
	return filepath.Join("/", t.RootDirHash(), t.FileName)
}

func (t *Text) genName(maintemplate string) string {
	file := filepath.Base(maintemplate)
	i := strings.LastIndex(file, ".")
	if i > -1 {
		file = file[:i]
	}
	sufix := ".html"
	if t.Extension != "" {
		sufix = t.Extension
	}
	return file + sufix
}

// Name method returns the file name of the cached document.
func (t *Text) Name() string {
	return t.FileName
}

// Modification shows the time of modification of the doc.
func (t *Text) Modification() time.Time {
	return t.MTime
}

// Sulfix return the sulfix of the file.
func (t *Text) Sulfix() string {
	return t.Extension
}

// New returns a new doc with the same metadata but the mtime. The new mtime
// will come from the cached version of the doc.
func (t *Text) New(m time.Time) Document {
	return &Text{
		FileName:  t.FileName,
		MTime:     m,
		Lang:      t.Lang,
		Encoding:  t.Encoding,
		Mime:      t.Mime,
		UserAgent: t.UserAgent,
		Extension: t.Extension,
		Hash:      t.Hash,
	}
}

// OutDate tells if the cache is out of date in relation to the original
// text.
func (t *Text) OutDate(cached Document) bool {
	tc, ok := cached.(*Text)
	if !ok {
		return true
	}
	if t.MTime.After(tc.MTime) {
		return true
	}
	if t.Lang != tc.Lang {
		return true
	}
	if t.Encoding != tc.Encoding {
		return true
	}
	if t.Mime != tc.Mime {
		return true
	}
	if t.UserAgent != tc.UserAgent {
		return true
	}
	if t.Hash != tc.Hash {
		return true
	}
	return false
}

// RootDirHash make a hash that is used in the root dir. This root dir name will
// identify a group of documents with the same caracteristics.
func (t *Text) RootDirHash() string {
	data := []byte(t.Encoding + t.Lang + t.Mime + t.UserAgent + t.Hash)
	return fmt.Sprintf("%x", md5.Sum(data))
}

// Image represents the metadata, cache relevant information. The local version.
type Image struct {
	FileName  string
	MTime     time.Time
	Lang      string
	Mime      string
	UserAgent string
	//ResponseCode string
	Hash      string
	Width     int
	Height    int
	Extension string
}

// ComposeFileName creates the file name from the data find in the request.
func (i *Image) ComposeFileName(r *http.Request) {
	i.FileName = filepath.Join("/", dir(r.URL), i.genName(name(r.URL)))
}

func (i *Image) genName(maintemplate string) string {
	file := filepath.Base(maintemplate)
	idx := strings.LastIndex(file, ".")
	if idx > -1 {
		file = file[:idx]
	}
	sufix := ".image"
	if i.Extension != "" {
		sufix = i.Extension
	}
	return file + sufix
}

// Name method returns the file name of the cached document.
func (i *Image) Name() string {
	return i.FileName
}

// Path return the full path of the cache
func (i *Image) Path() string {
	return filepath.Join("/", i.RootDirHash(), i.FileName)
}

// Modification shows the time of modification of the doc.
func (i *Image) Modification() time.Time {
	return i.MTime
}

// Sulfix return the sulfix of the file.
func (i *Image) Sulfix() string {
	return i.Extension
}

// New returns a new doc with the same metadata but the mtime. The new mtime
// will come from the cached version of the doc.
func (i *Image) New(m time.Time) Document {
	return &Image{
		FileName:  i.FileName,
		MTime:     m,
		Lang:      i.Lang,
		Mime:      i.Mime,
		Hash:      i.Hash,
		Width:     i.Width,
		Height:    i.Height,
		UserAgent: i.UserAgent,
		Extension: i.Extension,
	}
}

// OutDate tells if the cache is out of date in relation to  the original
// image.
func (i *Image) OutDate(cached Document) bool {
	ic, ok := cached.(*Image)
	if !ok {
		return true
	}
	if i.MTime.After(ic.MTime) {
		return true
	}
	if i.Lang != ic.Lang {
		return true
	}
	if i.Mime != ic.Mime {
		return true
	}
	if i.Hash != ic.Hash {
		return true
	}
	if i.Width != ic.Width {
		return true
	}
	if i.Height != ic.Height {
		return true
	}
	if i.UserAgent != ic.UserAgent {
		return true
	}
	return false
}

// RootDirHash make a hash that is used in the root dir. This root dir name will
// identify a group of documents with the same caracteristics.
func (i *Image) RootDirHash() string {
	data := []byte(i.Lang + i.Mime + strconv.FormatInt(int64(i.Width), 10) + strconv.FormatInt(int64(i.Height), 10) + i.UserAgent + i.Hash)
	return fmt.Sprintf("%x", md5.Sum(data))
}

func dir(u *url.URL) string {
	return filepath.Dir("/" + u.Path)
}

func name(u *url.URL) string {
	return filepath.Base("/" + u.Path)
}
