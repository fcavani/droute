// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package basicauth

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/fcavani/e"
)

func TestBasic(t *testing.T) {
	r, err := http.NewRequest("200", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = Basic(r)
	if err != nil && !e.Equal(err, ErrNoAuthHeader) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error nil")
	}

	r, err = http.NewRequest("200", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header["Authorization"] = []string{}
	_, _, err = Basic(r)
	if err != nil && !e.Equal(err, ErrInvalidAuthHeader) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error nil")
	}

	r, err = http.NewRequest("200", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Authorization", "teste")
	_, _, err = Basic(r)
	if err != nil && !e.Equal(err, ErrInvalidAuthString) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error nil")
	}

	r, err = http.NewRequest("200", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Authorization", "Complex teste")
	_, _, err = Basic(r)
	if err != nil && !e.Equal(err, ErrInvalidAuthScheme) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error nil")
	}

	r, err = http.NewRequest("200", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Authorization", "Basic teste:pass:seila")
	_, _, err = Basic(r)
	if err != nil && !e.Equal(err, ErrDecodeUsrPass) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error nil")
	}

	r, err = http.NewRequest("200", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("teste:pass:seila")))
	_, _, err = Basic(r)
	if err != nil && !e.Equal(err, ErrInvalidUsrPassString) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error nil")
	}

	r, err = http.NewRequest("200", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("teste:pass")))
	usr, pass, err := Basic(r)
	if err != nil {
		t.Fatal(err)
	}
	if usr != "teste" {
		t.Fatal("wrong user name")
	}
	if pass != "pass" {
		t.Fatal("wrong pass")
	}
}
