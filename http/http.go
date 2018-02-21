// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package http

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/fcavani/e"
	log "github.com/fcavani/slog"
)

// HTTPServer is a http and https server.
type HTTPServer struct {
	HTTPAddr           string
	HTTPSAddr          string
	Certificate        string
	PrivateKey         string
	CA                 string
	InsecureSkipVerify bool

	Handler http.Handler

	lnHTTP  net.Listener
	lnHTTPS net.Listener
}

// Init initializes the server.
func (h *HTTPServer) Init() error {
	var err error
	// Start the listners, open the doors
	// http
	log.Tag("router").Println("Setup http server...")
	h.lnHTTP, err = net.Listen("tcp", h.HTTPAddr)
	if err != nil {
		return e.Forward(err)
	}
	go func() {
		err = http.Serve(h.lnHTTP, h.Handler)
		if err != nil {
			log.Tag("router").Errorln("ListenAndServe failed:", err)
		}
	}()
	//https
	if !(h.Certificate == "" || h.PrivateKey == "") {
		log.Tag("router").Println("Setup https server...")
		certs := make([]tls.Certificate, 1)
		certs[0], err = tls.LoadX509KeyPair(
			h.Certificate,
			h.PrivateKey,
		)
		if err != nil {
			return e.Push(err, "LoadX509KeyPair failed")
		}
		var CAPool *x509.CertPool
		if h.CA != "" {
			CAPool = x509.NewCertPool()
			severCA, err := ioutil.ReadFile(h.CA)
			if err != nil {
				return e.Push(err, "could not load server CA")
			}
			CAPool.AppendCertsFromPEM(severCA)
		}
		httpsServer := &http.Server{
			Addr: h.HTTPSAddr,
			TLSConfig: &tls.Config{
				InsecureSkipVerify: h.InsecureSkipVerify,
				RootCAs:            CAPool,
				Certificates:       certs,
			},
			Handler: h.Handler,
		}
		httpsServer.TLSConfig.BuildNameToCertificate()
		conn, err := net.Listen("tcp", h.HTTPSAddr)
		if err != nil {
			return e.New(err)
		}
		h.lnHTTPS = tls.NewListener(conn, httpsServer.TLSConfig)
		if h.lnHTTPS == nil {
			return e.New("can't start tls listener")
		}
		go func() {
			err = httpsServer.Serve(h.lnHTTPS)
			if err != nil {
				log.Tag("router").Errorln("Server failed to start:", err)
			}
		}()
	}
	return nil
}

// Stop halts the router and close the listners.
func (h *HTTPServer) Stop() error {
	err := h.lnHTTP.Close()
	if err != nil {
		return e.Forward(err)
	}
	err = h.lnHTTPS.Close()
	if err != nil {
		return e.Forward(err)
	}
	return nil
}

// GetHTTPAddr get the bind address of the http server.
func (h *HTTPServer) GetHTTPAddr() string {
	return h.lnHTTP.Addr().String()
}

// GetHTTPSAddr get the bind address of the https server.
func (h *HTTPServer) GetHTTPSAddr() string {
	return h.lnHTTPS.Addr().String()
}
