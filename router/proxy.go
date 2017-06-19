// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fcavani/e"
	"github.com/fcavani/log"

	"github.com/fcavani/droute/errhandler"
	"github.com/fcavani/droute/responsewriter"
)

// HTTPClient is the default client to contact the server.
var HTTPClient *http.Client

func init() {
	HTTPClient = http.DefaultClient
}

// ConfigProxyHTTPClient configure one client with specific ca and insecure flag.
func ConfigProxyHTTPClient(ca string, insecure bool) error {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecure,
	}
	caCert, err := ioutil.ReadFile(ca)
	if err != nil {
		return e.Push(err, e.New("invalid root ca"))
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool
	tlsConfig.BuildNameToCertificate()

	HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	return nil
}

// Proxy forward the requests coming on path to dst url.
func Proxy(path string, timeout time.Duration) responsewriter.HandlerFunc {
	return func(w *responsewriter.ResponseWriter, r *http.Request) {
		dst := r.Context().Value("proxyredirdst").(string)

		parsed, err := url.Parse(dst)
		if err != nil {
			err = e.Push(err, e.New("invalid destiny url"))
			log.Tag("router", "server", "proxy").Error(err)
			errhandler.ErrHandler(w, http.StatusInternalServerError, err)
			return
		}

		uurl, err := userURL(r)
		if err != nil {
			err = e.Push(err, e.New("can't parse the url in request."))
			log.Tag("router", "server", "proxy").Error(err)
			errhandler.ErrHandler(w, http.StatusInternalServerError, err)
			return
		}

		oldurl := uurl.String()
		r.Host = parsed.Host
		r.URL.Host = parsed.Host
		r.URL.Path = strings.TrimPrefix(uurl.Path, path)
		r.URL.Scheme = parsed.Scheme
		r.RequestURI = ""

		signal := make(chan error)

		go func() {
			resp, err := HTTPClient.Do(r)
			if err != nil {
				err = e.Push(err, e.New("can't forward the request."))
				log.Tag("router", "server", "proxy").Error(err)
				signal <- err
				return
			}
			defer resp.Body.Close()

			headers := w.Header()
			for k, vals := range resp.Header {
				for _, val := range vals {
					headers.Add(k, val)
				}
			}
			w.WriteHeader(resp.StatusCode)

			n, err := io.Copy(w, resp.Body)
			if err != nil {
				log.Tag("router", "server", "proxy").Error("Can't copy the buffer:", err)
				signal <- err
				return
			}
			log.Tag("router", "server", "proxy").Printf("%v => %v, %v bytes, %v (%v)", oldurl, r.URL, n, r.Method, resp.StatusCode)
			signal <- nil
		}()

		select {
		case err := <-signal:
			if err != nil {
				errhandler.ErrHandler(w, http.StatusInternalServerError, err)
			}
		case <-time.After(timeout):
			errhandler.ErrHandler(w, http.StatusRequestTimeout, e.New("proxy request timeout"))
		}
	}
}
