// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package cache

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/fcavani/afero"
	"github.com/fcavani/e"
	"gopkg.in/fcavani/httprouter.v2"

	"github.com/fcavani/droute/middlewares/request"
)

// MetaFS function retrives the modification time os the file being accessed.
func MetaFS(fs afero.Fs, prefix string) Meta {
	return func(r *http.Request) (Document, *http.Request, error) {
		req := request.Request(r)
		path := req.URL().Path
		lang := httprouter.ContentLang(r)
		if lang != "" {
			p := strings.TrimPrefix(path, "/"+lang)
			if len(p) < len(path) {
				path = p
			}
		}
		up, err := url.PathUnescape(path)
		if err != nil {
			return nil, nil, e.Forward(err)
		}
		//path := filepath.Join(dir, strings.TrimPrefix(up, prefix))
		cleanpath := filepath.Clean(strings.TrimPrefix(up, prefix))
		fi, err := fs.Stat(cleanpath)
		if err != nil {
			return nil, nil, e.Forward(err)
		}

		doc := &Text{
			MTime:     fi.ModTime(),
			Lang:      req.Language(),
			Encoding:  req.Encoding(),
			Mime:      "",
			UserAgent: req.Useragent().String(),
			Hash:      "",
			Extension: filepath.Ext(cleanpath),
		}

		doc.ComposeFileName(r)

		return doc, r, nil
	}
}
