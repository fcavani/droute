// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package compress

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"io"
	"net/http"
	"strings"

	"github.com/fcavani/e"

	"github.com/fcavani/droute/errhandler"
	"github.com/fcavani/droute/middlewares/request"
	"github.com/fcavani/droute/responsewriter"
)

// Compress midleware, compres the data by the request of the user.
func Compress(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		resp := responsewriter.NewResponseWriter()
		f(resp, r)
		if !doCompress(resp) {
			w.Header().Set("Content-Encoding", "identity")
			err = resp.Copy(w)
			if err != nil {
				errhandler.ErrHandler(w, http.StatusInternalServerError, e.Forward(err))
				return
			}
			return
		}
		req := request.Request(r)
		switch req.Encoding() {
		case "gzip":
			w.Header().Set("Content-Encoding", "gzip")
			var nw *gzip.Writer
			nw, err = gzip.NewWriterLevel(w, 9)
			if err != nil {
				errhandler.ErrHandler(w, http.StatusInternalServerError, e.Forward(err))
				return
			}
			defer nw.Close()
			err = resp.Copy(newRespWriter(w, nw))
			if err != nil {
				errhandler.ErrHandler(w, http.StatusInternalServerError, e.Forward(err))
				return
			}
		case "compress":
			w.Header().Set("Content-Encoding", "compress")
			nwc := lzw.NewWriter(w, lzw.LSB, 8)
			defer nwc.Close()
			err = resp.Copy(newRespWriter(w, nwc))
			if err != nil {
				errhandler.ErrHandler(w, http.StatusInternalServerError, e.Forward(err))
				return
			}
		case "deflate":
			w.Header().Set("Content-Encoding", "deflate")
			var nw *flate.Writer
			nw, err = flate.NewWriter(w, flate.DefaultCompression)
			if err != nil {
				errhandler.ErrHandler(w, http.StatusInternalServerError, e.Forward(err))
				return
			}
			defer nw.Close()
			err = resp.Copy(newRespWriter(w, nw))
			if err != nil {
				errhandler.ErrHandler(w, http.StatusInternalServerError, e.Forward(err))
				return
			}
		case "identity":
			w.Header().Set("Content-Encoding", "identity")
			err = resp.Copy(w)
			if err != nil {
				errhandler.ErrHandler(w, http.StatusInternalServerError, e.Forward(err))
				return
			}
		default:
			errhandler.ErrHandler(w, http.StatusInternalServerError, e.New("invalid compression algorithm"))
			return
		}
		return
	}
}

func doCompress(w http.ResponseWriter) bool {
	mime, _, err := contentType(w.Header())
	if err != nil {
		return false
	}
	typ, sub := splitType(mime)
	switch typ {
	case "application":
		switch sub {
		case "javascript", "json", "rss+xml", "rss+atom":
			return true
		}
	case "text":
		switch sub {
		case "html", "javascript", "css", "markdown", "csv", "xml":
			return true
		}
	case "image":
		switch sub {
		case "bmp":
			return true
		}
	}
	return false
}

type response struct {
	http.ResponseWriter
	io.Writer
}

func newRespWriter(wr http.ResponseWriter, nw io.Writer) *response {
	return &response{
		ResponseWriter: wr,
		Writer:         nw,
	}
}

func (r *response) Write(p []byte) (int, error) {
	return r.Writer.Write(p)
}

// ErrNoContentType is a error
const ErrNoContentType = "no Content-Type found in header"

// ErrInvContentType is a error
const ErrInvContentType = "invalid Content-Type"

func contentType(h http.Header) (mime, encoding string, err error) {
	content := h.Get("Content-Type")
	if content == "" {
		return "", "", e.New(ErrNoContentType)
	}
	s := strings.Split(content, ";")
	if len(s) > 2 || len(s) < 1 {
		return "", "", e.New(ErrInvContentType)
	}
	if len(s) == 2 {
		return strings.TrimSpace(s[0]), strings.TrimSpace(s[1]), nil
	}
	return s[0], "", nil
}

func splitType(mime string) (string, string) {
	s := strings.Split(mime, "/")
	if len(s) != 2 {
		return "", ""
	}
	return s[0], s[1]
}
