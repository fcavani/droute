// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package compress

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/fcavani/droute/middlewares/request"
	"github.com/fcavani/droute/responsewriter"
)

func TestCompress(t *testing.T) {
	reqs := []func(h func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request){
		func(h func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				r.Header.Add("Accept-Encoding", "identity")
				h(w, r)
			}
		},
		func(h func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				r.Header.Add("Accept-Encoding", "identity")
				h(w, r)
			}
		},
		func(h func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				r.Header.Add("Accept-Encoding", "identity")
				h(w, r)
			}
		},
		func(h func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				r.Header.Add("Accept-Encoding", "gzip")
				h(w, r)
			}
		},
		func(h func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				r.Header.Add("Accept-Encoding", "compress")
				h(w, r)
			}
		},
		func(h func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				r.Header.Add("Accept-Encoding", "deflate")
				h(w, r)
			}
		},
		func(h func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				r.Header.Add("Accept-Encoding", "bogomogo")
				h(w, r)
			}
		},
	}
	funcs := []func(w http.ResponseWriter, r *http.Request){
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "image/jpg")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("oi"))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("oi"))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "image/bmp")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("oi"))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "image/bmp")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("oi"))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("oi"))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("oi"))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("oi"))
		},
	}
	for i, f := range funcs {
		var err error
		var buf []byte
		var r *http.Request
		req := &request.Config{}
		*req = *request.DefaultConfig
		r, err = http.NewRequest("GET", "http://localhost", nil)
		if err != nil {
			t.Fatal(err)
		}
		w := responsewriter.NewResponseWriter()
		h := reqs[i](request.Handler(req, 10240, Compress(f)))
		h(w, r)
		switch ce := w.Header().Get("Content-Encoding"); ce {
		case "identity":
			buf = w.Bytes()
		case "gzip":
			var gr *gzip.Reader
			gr, err = gzip.NewReader(w)
			if err != nil {
				t.Fatal(i, err)
			}
			buf, err = ioutil.ReadAll(gr)
			if err != nil {
				t.Fatal(i, err)
			}
			gr.Close()
		case "compress":
			lr := lzw.NewReader(w, lzw.LSB, 8)
			buf, err = ioutil.ReadAll(lr)
			if err != nil {
				t.Fatal(i, err)
			}
			lr.Close()
		case "deflate":
			fr := flate.NewReader(w)
			buf, err = ioutil.ReadAll(fr)
			if err != nil {
				t.Fatal(i, err)
			}
			fr.Close()
		default:
			if i == 6 {
				t.Log("bogomogo", ce)
				continue
			}
			t.Fatal(i, "invalid content-encoding")
		}
		str := string(buf)
		if str != "oi" {
			t.Fatal(i, "invalid response", str)
		}
	}
}
