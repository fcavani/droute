// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package router

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/fcavani/e"
	fhttp "github.com/fcavani/http"
	"github.com/fcavani/httprouter"
	"github.com/fcavani/log"
	"github.com/fcavani/text"
	"github.com/sony/gobreaker"

	"github.com/fcavani/droute/errhandler"
	"github.com/fcavani/droute/responsewriter"
)

// DefaultRouter is the name of the default route for the router.
const DefaultRouter = "_def_"

// Routers are the named routers avaliable.
type Routers map[string]*httprouter.Router

//NewRouters creates a new group of routers
func NewRouters() Routers {
	r := make(Routers)
	r.Set(DefaultRouter, httprouter.New())
	return r
}

// Set add a new router with name to the group. If called again with the same
// name it will replace that router.
func (r Routers) Set(name string, router *httprouter.Router) {
	r[name] = router
}

//Del remove a router.
func (r Routers) Del(name string) {
	delete(r, name)
}

//Get get a named route or if it doesn't exist get the default one.
func (r Routers) Get(name string) *httprouter.Router {
	route, found := r[name]
	if found {
		return route
	}
	route, found = r[DefaultRouter]
	if found {
		return route
	}
	return nil
}

// Router implements a router for http requests received in a build in http
// server.
type Router struct {
	proxyTimeout time.Duration
	proxyRetries int

	lb LoadBalance

	hostSwitch HostSwitch

	routers     Routers
	handler     http.Handler
	middlewares func(last responsewriter.HandlerFunc) responsewriter.HandlerFunc
	cbs         map[string]*gobreaker.CircuitBreaker
}

// HTTPHandlers plugs toggeder the handlers.
func (r *Router) HTTPHandlers(f func(first http.Handler) http.Handler) {
	r.handler = f(r.hostSwitch)
}

// Middlewares sets the chain of middlers used by the router handlers.
// Parameter last must be part of the chain and must be the last middleware.
func (r *Router) Middlewares(f func(last responsewriter.HandlerFunc) responsewriter.HandlerFunc) {
	r.middlewares = f
}

// ServeHTTP satisfy the interface for this work like a http server.
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(rw, req)
}

//Start listners
func (r *Router) Start(routers Routers, lb LoadBalance, to time.Duration, proxyRetries int) error {
	r.proxyTimeout = to
	r.proxyRetries = proxyRetries
	r.routers = routers
	r.lb = lb
	r.hostSwitch = make(HostSwitch)
	defRouter := r.routers.Get(DefaultRouter)
	if defRouter == nil {
		return e.Forward("no default router")
	}
	r.handler = r.hostSwitch
	r.middlewares = func(last responsewriter.HandlerFunc) responsewriter.HandlerFunc {
		return last
	}

	r.hostSwitch.Set("localhost", defRouter)

	r.cbs = make(map[string]*gobreaker.CircuitBreaker)

	// Add internal routes to the endpoints for adding new routes by the remote
	// client.
	r.routes()
	return nil
}

// SetHTTPAddr sets the default route for the adderess of the http server.
func (r *Router) SetHTTPAddr(addr string) {
	defRouter := r.routers.Get(DefaultRouter)
	r.hostSwitch.Set(addr, defRouter)
}

// SetHTTPSAddr sets the default route for the adderess of the https server.
func (r *Router) SetHTTPSAddr(addr string) {
	r.SetHTTPAddr(addr)
}

// SetHostSwitch sets the router to one hostname.
func (r *Router) SetHostSwitch(domain, routername string) error {
	router := r.routers.Get(routername)
	if router == nil {
		return e.New("no router with this name found")
	}
	r.hostSwitch.Set(domain, router)
	return nil
}

// Stop halts the router and close the listners.
func (r *Router) Stop() error {
	return nil
}

// Add a new handler to domain. If routerName doesn't exist add route
// to the default router.
func (r *Router) Add(routerName, method, path, dst string) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		switch x := r.(type) {
		case error:
			err = x
		case string:
			err = e.New(x)
		default:
			err = e.New(fmt.Errorf("%v", x))
		}
	}()
	err = text.CheckLettersNumber(routerName, 2, 128)
	if err != nil && routerName != DefaultRouter {
		err = e.Push(err, "invalid route name")
		return
	}
	err = text.CheckLettersNumber(method, 3, 20)
	if err != nil {
		err = e.Push(err, "invalid method name")
		return
	}
	if path != "" {
		err = text.CheckFileName(path, 1, 128)
		if err != nil {
			err = e.Push(err, "invalid path name")
			return
		}
	} else {
		path = "/"
	}
	_, err = url.Parse(dst)
	if err != nil {
		err = e.Push(err, "invalid destiny host name")
		return
	}
	router := r.routers.Get(routerName)
	if router == nil {
		err = e.New("router not found")
		return
	}

	if h, _, _ := router.Lookup(method, path); h != nil {
		// if the method/path exist return, or only add the new address for the
		// new server.
		r.lb.AddAddrs(method, path, dst)
		return
	}

	router.Handle(method, path,
		responsewriter.Handler(
			r.middlewares(
				Retry(r.proxyRetries,
					Balance(r.lb, //route.Remove(method, path)
						CircuitBrake(r.cbs,
							Proxy("", r.proxyTimeout),
						),
					),
				),
			),
		),
	)
	r.lb.AddAddrs(method, path, dst)
	return
}

// func (r *Router) Del(routerName string) (err error) {
// 	router := r.routers.Get(routerName)
// 	if router == nil {
// 		err = e.New("router not found")
// 		return
// 	}
// 	router.
// }

// func (r *Router) Get(routerName string) (rs Routes, err error) {
// 		router.HandlerPaths(true, func(method string, path string, h http.HandlerFunc) bool {
//
// 		})
// }

// AddRoute adds a new route to the router. If routerName doesn't exist add route
// to the default router.
func (r *Router) AddRoute(routerName, method, path string, handler http.HandlerFunc) {
	err := text.CheckLettersNumber(routerName, 2, 128)
	if err != nil && routerName != DefaultRouter {
		panic("invalid route name")
	}
	err = text.CheckLettersNumber(method, 3, 20)
	if err != nil {
		panic("invalid method name")
	}
	if path != "" {
		err = text.CheckFileName(path, 1, 128)
		if err != nil {
			panic("invalid path name")
		}
	} else {
		path = "/"
	}
	router := r.routers.Get(routerName)
	if router == nil {
		panic("route not found")
	}
	router.Handle(method, path, handler)
}

// func (r *Router) DelRoute() {
//
// }

func (r *Router) routes() {
	def := r.routers[DefaultRouter]
	// Add a route.
	def.POST("/_router/add",
		localhost(
			addRoute(r),
		),
	)

	// Del a route where path is the url scaped path to the route.
	// def.DELETE("/_router/del/:hostname/:path",
	// 	localhost(
	// 		delRoute(r),
	// 	),
	// )

	// Get return all routes.
	// def.GET("/_router/get",
	// 	localhost(
	// 		getRoute(r),
	// 	),
	// )
}

// Op is a operation in the router.
type Op string

const (
	//RouteOpAdd is a op of type add.
	RouteOpAdd Op = "add"
	//RouteOpDel is a op of type del.
	RouteOpDel Op = "delete"
	//RouteOpGet is a op of type get.
	RouteOpGet Op = "get"
)

// Routes discrible a group of routes.
type Routes []*Route

// Route define one route.
type Route struct {
	Methode string
	Router  string
	Path    string
	RedirTo string
}

// BodyLimitSize is the max request body size.
var BodyLimitSize int64 = 1048576

func addRoute(r *Router) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var route Route
		body, err := ioutil.ReadAll(io.LimitReader(req.Body, BodyLimitSize))
		if err != nil {
			respError(w, http.StatusInternalServerError, err.Error(), RouteOpAdd)
			return
		}
		err = req.Body.Close()
		if err != nil {
			respError(w, http.StatusInternalServerError, err.Error(), RouteOpAdd)
			return
		}
		err = json.Unmarshal(body, &route)
		if err != nil {
			respError(w, http.StatusInternalServerError, err.Error(), RouteOpAdd)
			return
		}
		err = r.Add(route.Router, route.Methode, route.Path, route.RedirTo)
		if err != nil {
			response(
				w,
				422, // unprocessable entity
				route.Methode,
				route.Router,
				route.Path,
				err.Error(),
				RouteOpAdd,
			)
			return
		}
		response(
			w,
			http.StatusCreated,
			route.Methode,
			route.Router,
			route.Path,
			"",
			RouteOpAdd,
		)
	}
}

// func delRoute(r *Router) func(w http.ResponseWriter, req *http.Request) {
// 	return func(w http.ResponseWriter, req *http.Request) {
//    params := httprouter.Parameters(r)
// 		hn := params.ByName("hostname")
// 		p := params.ByName("path")
//
// 	}
// }

// func getRoute(r *Router) func(w http.ResponseWriter, req *http.Request) {
// 	return func(w http.ResponseWriter, req *http.Request) {
//
// 	}
// }

//OpErr handle the error to the client.
type OpErr struct {
	Err string `json:"err"`
	Op  string `json:"op"`
}

func (oe *OpErr) Error() string {
	if oe.Err == "" {
		return "no error"
	}
	return fmt.Sprintf(
		"op: %v, err: %v",
		oe.Op,
		oe.Err,
	)
}

func respError(w http.ResponseWriter, code int, err string, op Op) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	resp := OpErr{
		Err: err,
		Op:  string(op),
	}
	er := json.NewEncoder(w).Encode(resp)
	if er != nil {
		log.Tag("router", "server", "rest").Error(er)
	}
}

// Response is the response sent to the client
type Response struct {
	Methode string `json:"methode"`
	Router  string `json:"router"`
	Path    string `json:"path"`
	Err     string `json:"err"`
	Op      string `json:"op"`
}

func (r *Response) Error() string {
	if r.Err == "" {
		return "no error"
	}
	return fmt.Sprintf(
		"router: %v, method: %v, path %v, op: %v, error: %v",
		r.Router,
		r.Methode,
		r.Path,
		r.Op,
		r.Err,
	)
}

func response(w http.ResponseWriter, code int, methode, routerName, path, err string, op Op) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	resp := Response{
		Methode: methode,
		Router:  routerName,
		Path:    path,
		Err:     err,
		Op:      string(op),
	}
	er := json.NewEncoder(w).Encode(resp)
	if er != nil {
		log.Tag("router", "server", "rest").Error(er)
	}
}

func localhost(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ipstr, err := fhttp.RemoteIP(r)
		if err != nil {
			errhandler.ErrHandler(w, http.StatusForbidden, err)
			return
		}
		ip := net.ParseIP(ipstr)
		if ip.IsLoopback() {
			f(w, r)
			return
		}
		errhandler.ErrHandler(w, http.StatusForbidden, e.New("ip isn't loopback"))
	}
}
