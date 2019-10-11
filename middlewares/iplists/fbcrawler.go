// Copyright 2017 Felipe A. Cavani. All rights reserved.

package iplists

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/fcavani/e"
	log "github.com/fcavani/slog"
)

// Object is the whois AS object
type Object struct {
	Route   *net.IPNet
	Descr   string
	Origin  string
	MntBy   string
	Changed string
	Source  string
}

// FBCrawler contains the information about the crawlers.
type FBCrawler struct {
	Server  string
	Port    string
	AS      []string
	Refresh time.Duration
	Cache   string
	objs    map[string]*Object //cache
	lck     sync.Mutex
}

// NewFBCrawler create a new crawler query.
func NewFBCrawler(server, port string, refresh time.Duration, cache string, AS ...string) (*FBCrawler, error) {
	fbc := &FBCrawler{
		Server:  server,
		Port:    port,
		AS:      AS,
		Refresh: refresh,
		Cache:   cache,
		objs:    make(map[string]*Object),
	}
	err := fbc.init()
	if err != nil {
		return nil, e.Forward(err)
	}
	return fbc, nil
}

func (fbc *FBCrawler) init() error {
	var err error
	var f *os.File
	cache := make(map[string]*Object)

	f, err = os.Open(fbc.Cache)
	if err != nil {
		goto retrive
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&cache)
	if err != nil {
		goto retrive
	}
	fbc.objs = cache
	goto start
retrive:
	// Buscar po uma nova
	err = fbc.retrieve()
	if err != nil {
		log.Tag("fbcrawler").Println("Erro retriving the whois data:", e.Forward(err))
	}
	goto start
start:
	go func() {
		for {
			time.Sleep(fbc.Refresh)
			// TODO: QUEM ATUALIZA?????
			// fbc.retrieve()

		}
	}()
	return nil
}

func (fbc *FBCrawler) retrieve() error {
	for _, as := range fbc.AS {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(fbc.Server, fbc.Port), 15*time.Second)
		if err != nil {
			return e.Forward(err)
		}
		req := "-i origin " + as + "\r\n"
		_, err = conn.Write([]byte(req))
		if err != nil {
			conn.Close()
			return e.Forward(err)
		}
		buf := bytes.NewBuffer([]byte{})
		_, err = io.Copy(buf, conn)
		if err != nil {
			conn.Close()
			return e.Forward(err)
		}
		err = fbc.parse(buf)
		if err != nil {
			conn.Close()
			return e.Forward(err)
		}
		err = conn.Close()
		if err != nil {
			return e.Forward(err)
		}
	}
	//cache
	err := fbc.store()
	if err != nil {
		return e.Forward(err)
	}
	return nil
}

func (fbc *FBCrawler) store() error {
	f, err := os.OpenFile(fbc.Cache, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return e.Forward(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	fbc.lck.Lock()
	defer fbc.lck.Unlock()
	err = enc.Encode(fbc.objs)
	if err != nil {
		return e.Forward(err)
	}
	return nil
}

// route:      204.15.20.0/22
// descr:      Facebook, Inc.
// origin:     AS32934
// mnt-by:     MAINT-AS32934
// changed:    callahan@facebook.com 20090608  #00:40:18Z
// source:     RADB

var lineStartRoute = []byte("route:")
var lineStartRoute6 = []byte("route6:")
var lineStartDescr = []byte("descr:")
var lineStartOrigin = []byte("origin:")
var lineStartMntBy = []byte("mnt-by:")
var lineStartChanged = []byte("changed:")
var lineStartSource = []byte("source:")

func (fbc *FBCrawler) parse(buf *bytes.Buffer) error {
	var obj *Object
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return e.Forward(err)
		}
		switch {
		case bytes.Contains(line, []byte("No entries found")):
			return e.New("query error")
		case bytes.HasPrefix(line, lineStartRoute):
			obj = &Object{}
			ipstr := string(bytes.TrimSpace(bytes.TrimPrefix(line, lineStartRoute)))
			_, IPNet, er := net.ParseCIDR(ipstr)
			if er != nil {
				return e.Forward(er)
			}
			obj.Route = IPNet
		case bytes.HasPrefix(line, lineStartRoute6):
			obj = &Object{}
			ipstr := string(bytes.TrimSpace(bytes.TrimPrefix(line, lineStartRoute6)))
			_, IPNet, er := net.ParseCIDR(ipstr)
			if er != nil {
				return e.Forward(er)
			}
			obj.Route = IPNet
		case bytes.HasPrefix(line, lineStartDescr):
			obj.Descr = string(bytes.TrimSpace(bytes.TrimPrefix(line, lineStartDescr)))
		case bytes.HasPrefix(line, lineStartOrigin):
			obj.Origin = string(bytes.TrimSpace(bytes.TrimPrefix(line, lineStartOrigin)))
		case bytes.HasPrefix(line, lineStartMntBy):
			obj.MntBy = string(bytes.TrimSpace(bytes.TrimPrefix(line, lineStartMntBy)))
		case bytes.HasPrefix(line, lineStartChanged):
			obj.Changed = string(bytes.TrimSpace(bytes.TrimPrefix(line, lineStartChanged)))
		case bytes.HasPrefix(line, lineStartSource):
			obj.Source = string(bytes.TrimSpace(bytes.TrimPrefix(line, lineStartSource)))
		case bytes.Equal(line, []byte("\n")):
			if obj.Route == nil {
				return e.New("invalid ip addrs")
			}
			fbc.lck.Lock()
			fbc.objs[obj.Route.String()] = obj
			fbc.lck.Unlock()
			// default:
			// 	fmt.Println(string(line))
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}

// HaveIP return true if ip is a crawler.
func (fbc *FBCrawler) HaveIP(ipstr string) bool {
	ip := net.ParseIP(ipstr)
	fbc.lck.Lock()
	defer fbc.lck.Unlock()
	for _, v := range fbc.objs {
		if v.Route.Contains(ip) {
			return true
		}
	}
	return false
}

const ErrIterStop = "stop iter"

func (fbc *FBCrawler) Iter(f func(string, *Object) error) error {
	var err error
	fbc.lck.Lock()
	defer fbc.lck.Unlock()
	for k, v := range fbc.objs {
		err = f(k, v)
		if e.Equal(err, ErrIterStop) {
			return nil
		} else if err != nil {
			return e.Forward(err)
		}
	}
	return nil
}
