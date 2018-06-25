// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/fcavani/droute/router"
	"github.com/fcavani/e"
	neturl "github.com/fcavani/net/url"
	log "github.com/fcavani/slog"
	"gopkg.in/fcavani/httprouter.v2"
)

// Router agregate the methods to interact with the router server and add methods
// to it.
type Router struct {
	// Router name that will handle the request.
	Router string
	// URL of the router server. To access the REST service.
	URL *url.URL

	// Addrs of the host that code will be running.
	Addrs string

	router *httprouter.Router

	routes map[*router.Route]http.HandlerFunc
	lck    sync.Mutex
	once   sync.Once
}

//HTTPClient is the http.Client
var HTTPClient *http.Client

func init() {
	HTTPClient = http.DefaultClient
}

// ConfigHTTPClient config the http client that is used by do method to access
// the router.
func ConfigHTTPClient(certificate, privateKey, ca string, insecure bool) error {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecure,
	}
	tlsConfig.Certificates = []tls.Certificate{}
	// Load client cert
	if !(certificate == "" || privateKey == "") {
		cert, err := tls.LoadX509KeyPair(certificate, privateKey)
		if err != nil {
			return e.New(err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if ca != "" {
		// Load CA cert
		caCert, err := ioutil.ReadFile(ca)
		if err != nil {
			return e.New(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	// Setup HTTPS client
	tlsConfig.BuildNameToCertificate()

	HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	return nil
}

// Start initialize the router. Setup the server that will receive the income
// requests and send it to the right route.
func (r *Router) Start() error {
	r.once.Do(func() {
		r.router = httprouter.New()
		// TODO: Tirar isso daqui. Automatizar...
		r.router.DefaultLang = "pt"
		r.router.SupportedLangs = map[string]struct{}{
			"pt": struct{}{},
			"en": struct{}{},
		}
		r.routes = make(map[*router.Route]http.HandlerFunc)
		go func() {
			for {
				select {
				case <-time.After(time.Minute):
					r.lck.Lock()
					for route, handler := range r.routes {
						err := r.handlerfunc(route, handler)
						if err == nil {
							log.Errorf("Route re add to proxy: %v %v %v", route.Router, route.Methode, route.Path)
						}
					}
					r.lck.Unlock()
				}
			}
		}()
	})
	return nil
}

// ServeHTTP is a http server with the route setup by *Router.
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(rw, req)
}

// DELETE route with http method DELETE.
func (r *Router) DELETE(path string, handle http.HandlerFunc) error {
	return r.HandlerFunc("DELETE", path, handle)
}

// GET route with http method GET.
func (r *Router) GET(path string, handle http.HandlerFunc) error {
	return r.HandlerFunc("GET", path, handle)
}

// HEAD route with http method HEAD.
func (r *Router) HEAD(path string, handle http.HandlerFunc) error {
	return r.HandlerFunc("HEAD", path, handle)
}

// OPTIONS route with http method OPTIONS.
func (r *Router) OPTIONS(path string, handle http.HandlerFunc) error {
	return r.HandlerFunc("OPTIONS", path, handle)
}

// PATCH route with http method PATCH.
func (r *Router) PATCH(path string, handle http.HandlerFunc) error {
	return r.HandlerFunc("PATCH", path, handle)
}

// POST route with http method POST.
func (r *Router) POST(path string, handle http.HandlerFunc) error {
	return r.HandlerFunc("POST", path, handle)
}

// PUT route with http method PUT.
func (r *Router) PUT(path string, handle http.HandlerFunc) error {
	return r.HandlerFunc("PUT", path, handle)
}

// BodyLimitSize is the max request body size.
var BodyLimitSize int64 = 1048576

// HandlerFunc is a generic method to add a route into the router.
func (r *Router) HandlerFunc(method, path string, handler http.HandlerFunc) error {
	r.lck.Lock()
	defer r.lck.Unlock()

	route := &router.Route{
		Methode: method,
		Router:  r.Router,
		Path:    path,
		RedirTo: r.Addrs,
	}

	err := r.handlerfunc(route, handler)
	if err != nil {
		return err
	}

	r.routes[route] = handler

	return nil
}

func (r *Router) handlerfunc(route *router.Route, handler http.HandlerFunc) (err error) {
	var body []byte

	defer func() {
		if err != nil {
			log.Errorf("Can't add handler (%v, %v, %v) error: %v", route.Router, route.Methode, route.Path, err)
		}
	}()
	buf, err := json.Marshal(route)
	if err != nil {
		err = e.Forward(err)
		return
	}
	u := neturl.Copy(r.URL)
	u.Path = "/en/_router/add"
	resp, err := HTTPClient.Post(u.String(), "application/json", bytes.NewReader(buf))
	if err != nil {
		err = e.Forward(err)
		return
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusCreated:
		defer func() {
			if err != nil {
				return
			}
			r := recover()
			switch x := r.(type) {
			case error:
				err = x
			case string:
				err = e.New(x)
			default:
				if x != nil {
					err = e.New(x)
				}
			}
		}()
		r.router.Handle(route.Methode, route.Path, handler)
		return
	case 422:
		response := &router.Response{}
		body, err = ioutil.ReadAll(io.LimitReader(resp.Body, BodyLimitSize))
		if err != nil {
			err = e.Forward(err)
			return
		}
		err = json.Unmarshal(body, response)
		if err != nil {
			err = e.Forward(err)
			return
		}
		return response
	case http.StatusInternalServerError:
		operr := &router.OpErr{}
		body, err = ioutil.ReadAll(io.LimitReader(resp.Body, BodyLimitSize))
		if err != nil {
			err = e.Forward(err)
			return
		}
		err = json.Unmarshal(body, operr)
		if err != nil {
			err = e.Forward(err)
			return
		}
		return operr
	default:
		err = e.New("failed to add a function handler to the router. (status code %v)", resp.StatusCode)
		return
	}
}
