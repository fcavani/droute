// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/fcavani/e"
	"github.com/fcavani/log"
	"gopkg.in/fcavani/httprouter.v2"

	drouterhttp "github.com/fcavani/droute/http"
	"github.com/fcavani/droute/router"
)

var dr *router.Router
var drh *drouterhttp.HTTPServer
var clientRouter *Router
var h *drouterhttp.HTTPServer

func TestRouter(t *testing.T) {
	dr = new(router.Router)
	routers := router.NewRouters()
	def := routers.Get(router.DefaultRouter)
	def.Context = func(ctx context.Context) context.Context {
		ctx, _ = router.WithSignal(ctx, os.Interrupt, os.Kill)
		ctx, _ = context.WithDeadline(ctx, time.Now().Add(100*time.Millisecond))
		return ctx
	}
	err := dr.Start(
		routers,
		router.NewRoundRobin(),
		5*time.Second, //timeout proxy
		3,             // retry proxy
	)
	if err != nil {
		t.Fatal(err)
	}
	err = router.ConfigProxyHTTPClient("../rootCA.pem", true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRouterHTTP(t *testing.T) {
	drh = &drouterhttp.HTTPServer{
		HTTPAddr:           "localhost:8081",
		HTTPSAddr:          "localhost:8082",
		Certificate:        "../device.crt",
		PrivateKey:         "../device.key",
		CA:                 "../rootCA.pem",
		InsecureSkipVerify: true,
		Handler:            dr,
	}

	err := drh.Init()
	if err != nil {
		t.Fatal(err)
	}

	dr.SetHTTPAddr(drh.GetHTTPAddr())
	dr.SetHTTPSAddr(drh.GetHTTPSAddr())

	dr.SetHostSwitch("localhost:8081", router.DefaultRouter)
	dr.SetHostSwitch("localhost:8082", router.DefaultRouter)
}

func TestClientRouter(t *testing.T) {
	u, _ := url.Parse("https://localhost:8082/")

	err := ConfigHTTPClient("", "", "../rootCA.pem", true)
	if err != nil {
		t.Fatal(err)
	}

	clientRouter = &Router{
		Router: router.DefaultRouter,
		URL:    u,
		Addrs:  "https://localhost:8084",
	}

	err = clientRouter.Start()
	if err != nil {
		t.Fatal(err)
	}
}

// func TestClientRouterClose(t *testing.T) {
//   err := clientRouter.
// }

func TestClientHTTP(t *testing.T) {
	h = &drouterhttp.HTTPServer{
		HTTPAddr:           "localhost:8083",
		HTTPSAddr:          "localhost:8084",
		Certificate:        "../device.crt",
		PrivateKey:         "../device.key",
		CA:                 "../rootCA.pem",
		InsecureSkipVerify: true,
		Handler:            clientRouter,
	}

	err := h.Init()
	if err != nil {
		t.Fatal(err)
	}
}

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

const addrs = "https://localhost:8082/"

func TestGET(t *testing.T) {
	err := clientRouter.GET("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, "%v", "teste")
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := httpClient.Get(addrs)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("wrong status code,", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "teste" {
		t.Fatal("wrong response, body message invalid")
	}
}

func TestGET2(t *testing.T) {
	err := clientRouter.GET("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, "%v", "another test")
	})
	if err != nil && !e.Contains(err, "a handle is already registered for path") {
		t.Fatal(err)
	}
}

func TestGET3(t *testing.T) {
	err := clientRouter.GET("/:data/:entry", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, "%v", "teste")
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := httpClient.Get(addrs)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("wrong status code,", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "teste" {
		t.Fatal("wrong response, body message invalid")
	}
}

func TestDELETE(t *testing.T) {
	err := clientRouter.DELETE("/*filename", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		file := httprouter.Parameters(req).ByName("filename")
		fmt.Fprintf(rw, "file deleted %v", file)
	})
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("DELETE", addrs+"robots.txt", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := httpClient.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("wrong status code,", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "file deleted /robots.txt" {
		t.Fatal("wrong response, body message invalid", string(body))
	}
}

func TestHEAD(t *testing.T) {
	err := clientRouter.HEAD("/file.txt", func(rw http.ResponseWriter, req *http.Request) {
		expire := 600 * time.Second
		maxAge := strconv.FormatFloat(expire.Seconds(), 'f', 0, 64)
		rw.Header().Add("Content-Length", "69")
		rw.Header().Add("Cache-Control", "max-age="+maxAge+", public")
		rw.Header().Add("Expires", time.Now().Add(expire).UTC().Format(time.RFC1123))
		rw.Header().Add("Last-Modified", time.Now().UTC().Format(time.RFC1123))
		rw.Header().Add("Etag", "kLsKpOO0K")
		rw.WriteHeader(http.StatusOK)
	})
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("HEAD", addrs+"file.txt", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := httpClient.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("wrong status code,", resp.StatusCode)
	}
	if resp.Header.Get("Etag") != "kLsKpOO0K" {
		t.Fatal("response header is invalid")
	}
	if resp.Header.Get("Content-Length") != "69" {
		t.Fatal("response header is invalid")
	}
}

func TestOPTIONS(t *testing.T) {
	err := clientRouter.OPTIONS("/catoto.dat", func(rw http.ResponseWriter, req *http.Request) {
		expire := 600 * time.Second
		maxAge := strconv.FormatFloat(expire.Seconds(), 'f', 0, 64)
		rw.Header().Add("Content-Length", "0")
		rw.Header().Add("Cache-Control", "max-age="+maxAge+", public")
		rw.Header().Add("Expires", time.Now().Add(expire).UTC().Format(time.RFC1123))
		rw.Header().Add("Last-Modified", time.Now().UTC().Format(time.RFC1123))
		rw.Header().Add("Allow", "GET, HEAD, OPTION")
		rw.Header().Add("Server", "no_server_name_v1")
		rw.WriteHeader(http.StatusOK)
	})
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("OPTIONS", addrs+"catoto.dat", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := httpClient.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("wrong status code,", resp.StatusCode)
	}
	if resp.Header.Get("Server") != "no_server_name_v1" {
		t.Fatal("response header is invalid")
	}
	if resp.Header.Get("Content-Length") != "0" {
		t.Fatal("response header is invalid")
	}
	if resp.Header.Get("Allow") != "GET, HEAD, OPTION" {
		t.Fatal("response header is invalid")
	}
}

func TestPATCH(t *testing.T) {
	err := clientRouter.PATCH("/user.json", func(rw http.ResponseWriter, req *http.Request) {
		var user struct{ Name string }
		er := json.NewDecoder(req.Body).Decode(&user)
		if er != nil {
			t.Log(er)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		if user.Name != "pafuncio" {
			t.Log("invalid name")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()
		etag := req.Header.Get("If-Match")
		if etag != "1" {
			t.Log("invalid If-Match")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		rw.Header().Add("Etag", "2")
		rw.Header().Add("Content-Location", "/user.json")
		rw.WriteHeader(http.StatusOK)
	})
	if err != nil {
		t.Fatal(err)
	}
	reqBody := bytes.NewBufferString("{ \"Name\": \"pafuncio\" }")
	r, err := http.NewRequest("PATCH", addrs+"user.json", reqBody)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("If-Match", "1")
	resp, err := httpClient.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("wrong status code,", resp.StatusCode)
	}
	if resp.Header.Get("Etag") != "2" {
		t.Fatal("response header is invalid")
	}
	if resp.Header.Get("Content-Location") != "/user.json" {
		t.Fatal("response header is invalid")
	}
}

func TestPOST(t *testing.T) {
	err := clientRouter.POST("/form", func(rw http.ResponseWriter, req *http.Request) {
		er := req.ParseForm()
		if er != nil {
			t.Log(er)
			rw.WriteHeader(http.StatusInternalServerError)
		}
		if req.Form.Get("field0") != "teste0" {
			t.Log("invalid form field")
			rw.WriteHeader(http.StatusInternalServerError)
		}
		if req.Form.Get("field1") != "teste1" {
			t.Log("invalid form field")
			rw.WriteHeader(http.StatusInternalServerError)
		}
		rw.WriteHeader(http.StatusOK)
	})
	if err != nil {
		t.Fatal(err)
	}

	vals := make(url.Values)
	vals.Add("field0", "teste0")
	vals.Add("field1", "teste1")

	resp, err := httpClient.PostForm(addrs+"/form", vals)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("wrong status code,", resp.StatusCode)
	}
}

func TestPUT(t *testing.T) {
	err := clientRouter.PUT("/*filename", func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != "text/html" {
			t.Log("invalid Content-Type")
			rw.WriteHeader(http.StatusInternalServerError)
		}
		rw.WriteHeader(http.StatusOK)
	})
	if err != nil {
		t.Fatal(err)
	}
	reqBody := bytes.NewBufferString("<p>new bla</p>")
	r, err := http.NewRequest("PUT", addrs+"/new.html", reqBody)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("Content-Type", "text/html")
	resp, err := httpClient.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("wrong status code,", resp.StatusCode)
	}
}

func Test422(t *testing.T) {
	r := httprouter.New()
	r.POST("/_router/add", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
		rw.WriteHeader(422)
		resp := router.Response{
			Methode: "GET",
			Router:  "routerName",
			Path:    "/",
			Err:     "dummy error",
			Op:      "ADD",
		}
		er := json.NewEncoder(rw).Encode(resp)
		if er != nil {
			log.Tag("router", "server", "rest").Error(er)
		}
	})
	server := &drouterhttp.HTTPServer{
		HTTPAddr:           "localhost:8085",
		HTTPSAddr:          "localhost:8086",
		Certificate:        "../device.crt",
		PrivateKey:         "../device.key",
		CA:                 "../rootCA.pem",
		InsecureSkipVerify: true,
		Handler:            r,
	}
	err := server.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()

	u, _ := url.Parse("https://localhost:8086/")
	cr := &Router{
		Router: router.DefaultRouter,
		URL:    u,
		Addrs:  "https://localhost:8088",
	}
	err = cr.Start()
	if err != nil {
		t.Fatal(err)
	}

	client := &drouterhttp.HTTPServer{
		HTTPAddr:           "localhost:8087",
		HTTPSAddr:          "localhost:8088",
		Certificate:        "../device.crt",
		PrivateKey:         "../device.key",
		CA:                 "../rootCA.pem",
		InsecureSkipVerify: true,
		Handler:            cr,
	}
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Stop()

	err = cr.GET("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, "%v", "teste")
	})
	if err != nil {
		if resp, ok := err.(*router.Response); ok {
			if resp.Err != "dummy error" {
				t.Fatal("invalid response")
			}
		} else {
			t.Fatal(err)
		}
	}

	resp, err := httpClient.Get("https://localhost:8086/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("wrong status code,", resp.StatusCode)
	}
}

func Test500(t *testing.T) {
	r := httprouter.New()
	r.POST("/_router/add", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
		rw.WriteHeader(500)
		resp := router.OpErr{
			Err: "some error",
			Op:  "ADD",
		}
		er := json.NewEncoder(rw).Encode(resp)
		if er != nil {
			log.Tag("router", "server", "rest").Error(er)
		}
	})
	server := &drouterhttp.HTTPServer{
		HTTPAddr:           "localhost:8005",
		HTTPSAddr:          "localhost:8006",
		Certificate:        "../device.crt",
		PrivateKey:         "../device.key",
		CA:                 "../rootCA.pem",
		InsecureSkipVerify: true,
		Handler:            r,
	}
	err := server.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()

	u, _ := url.Parse("https://localhost:8006/")
	cr := &Router{
		Router: router.DefaultRouter,
		URL:    u,
		Addrs:  "https://localhost:8008",
	}
	err = cr.Start()
	if err != nil {
		t.Fatal(err)
	}

	client := &drouterhttp.HTTPServer{
		HTTPAddr:           "localhost:8007",
		HTTPSAddr:          "localhost:8008",
		Certificate:        "../device.crt",
		PrivateKey:         "../device.key",
		CA:                 "../rootCA.pem",
		InsecureSkipVerify: true,
		Handler:            cr,
	}
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Stop()

	err = cr.GET("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, "%v", "teste")
	})
	if err != nil {
		if resp, ok := err.(*router.OpErr); ok {
			if resp.Err != "some error" {
				t.Fatal("invalid response")
			}
		} else {
			t.Fatal(err)
		}
	}

	resp, err := httpClient.Get("https://localhost:8006/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("wrong status code,", resp.StatusCode)
	}
}

func TestXXX(t *testing.T) {
	r := httprouter.New()
	r.POST("/_router/add", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
		rw.WriteHeader(404)
	})
	server := &drouterhttp.HTTPServer{
		HTTPAddr:           "localhost:8015",
		HTTPSAddr:          "localhost:8016",
		Certificate:        "../device.crt",
		PrivateKey:         "../device.key",
		CA:                 "../rootCA.pem",
		InsecureSkipVerify: true,
		Handler:            r,
	}
	err := server.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()

	u, _ := url.Parse("https://localhost:8016/")
	cr := &Router{
		Router: router.DefaultRouter,
		URL:    u,
		Addrs:  "https://localhost:8008",
	}
	err = cr.Start()
	if err != nil {
		t.Fatal(err)
	}

	client := &drouterhttp.HTTPServer{
		HTTPAddr:           "localhost:8017",
		HTTPSAddr:          "localhost:8018",
		Certificate:        "../device.crt",
		PrivateKey:         "../device.key",
		CA:                 "../rootCA.pem",
		InsecureSkipVerify: true,
		Handler:            cr,
	}
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Stop()

	err = cr.GET("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, "%v", "teste")
	})
	if err != nil && !e.Contains(err, "failed to add a function handler to the router") {
		t.Fatal(err)
	}

	resp, err := httpClient.Get("https://localhost:8016/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("wrong status code,", resp.StatusCode)
	}
}

func TestRouterClose(t *testing.T) {
	err := dr.Stop()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRouterHTTPClose(t *testing.T) {
	err := drh.Stop()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClientHTTPClose(t *testing.T) {
	err := h.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
