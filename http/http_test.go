// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package http

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/fcavani/httprouter"
)

var httpClient *http.Client

func TestHTTPClient(t *testing.T) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	caCert, err := ioutil.ReadFile("../rootCA.pem")
	if err != nil {
		t.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool
	tlsConfig.BuildNameToCertificate()

	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
}

func TestHTTP(t *testing.T) {
	router := httprouter.New()
	router.GET("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(200)
		fmt.Fprint(rw, "oi")
	})

	hs := &HTTPServer{
		HTTPAddr:           "localhost:0",
		HTTPSAddr:          "localhost:0",
		Certificate:        "../device.crt",
		PrivateKey:         "../device.key",
		CA:                 "../rootCA.pem",
		InsecureSkipVerify: true,
		Handler:            router,
	}

	err := hs.Init()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(hs.GetHTTPSAddr())

	resp, err := httpClient.Get("https://" + hs.GetHTTPSAddr())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("wrong status code,", resp.StatusCode)
	}

	err = hs.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
