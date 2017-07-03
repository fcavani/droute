// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package cache

import (
	"net/http"
	"testing"

	"github.com/fcavani/afero"
	"github.com/fcavani/e"

	"github.com/fcavani/droute/middlewares/request"
	"github.com/fcavani/droute/responsewriter"
)

func TestMetafs(t *testing.T) {
	mem := afero.NewMemMapFs()
	err := mem.MkdirAll("/this/is/not/a/dir/", 0770)
	if err != nil {
		t.Fatal(err)
	}
	base := afero.NewBasePathFs(mem, "/this/is/not/a/dir/")

	req, err := http.NewRequest("GET", "http://localhost/static/file01.dat", nil)
	if err != nil {
		t.Fatal(err)
	}

	rw := responsewriter.NewResponseWriter()

	request.Handler(request.DefaultConfig, 10240, func(rw http.ResponseWriter, r *http.Request) {
		*req = *r
	})(rw, req)

	meta := MetaFS(base, "/static/")
	_, _, err = meta(req)
	if err != nil && !e.Contains(err, "file does not exist") {
		t.Fatal(err)
	}

	f, err := base.Create("file01.dat")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	doc, _, err := meta(req)
	if err != nil {
		t.Fatal(err)
	}

	_ = doc
}
