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

	"github.com/fcavani/droute/middlewares/request"
)

// MetaFS function retrives the modification time os the file being accessed.
func MetaFS(fs afero.Fs, prefix string) Meta {
	return func(r *http.Request) (Document, *http.Request, error) {
		req := request.Request(r)
		up, err := url.PathUnescape(req.Url().Path)
		if err != nil {
			return nil, nil, e.Forward(err)
		}
		//path := filepath.Join(dir, strings.TrimPrefix(up, prefix))
		path := filepath.Clean(strings.TrimPrefix(up, prefix))
		fi, err := fs.Stat(path)
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
			Extension: filepath.Ext(path),
		}

		doc.ComposeFileName(r)

		return doc, r, nil
	}
}
