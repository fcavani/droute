// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package etcd

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff"

	etcdCli "github.com/coreos/etcd/client"
	"github.com/fcavani/e"
	log "github.com/fcavani/slog"
	"github.com/spf13/viper"
	"github.com/xordataexchange/crypt/config"
	"github.com/xordataexchange/crypt/encoding/secconf"
)

func LoadEtcdEndpoints(endpoints string) []string {
	eps := make([]string, 0)
	etcdConfs := viper.GetStringMapStringSlice("etcdCli")
	if etcdConfs != nil {
		endpoints := etcdConfs["endpoints"]
		if endpoints != nil {
			for _, endpoint := range endpoints {
				eps = append(eps, endpoint)
			}
		}
	}
	if endpoints != "" {
		s := strings.Split(endpoints, ",")
		for _, ep := range s {
			ep = strings.TrimSpace(ep)
			eps = append(eps, ep)
		}
	}

	if len(eps) == 0 {
		log.Fatal("No endpoints for etcdCli, can't do the configuration.")
	}

	return eps
}

type Etcd struct {
	Endpoints  []string
	SecKeyRing string
	kapi       etcdCli.KeysAPI
	cm         config.ConfigManager
	keystore   []byte
}

func (etc *Etcd) Init() error {
	var err error
	if len(etc.Endpoints) == 0 {
		return e.New("no end points")
	}
	cfg := etcdCli.Config{
		Endpoints: etc.Endpoints,
		//Transport: http.DefaultTransport,
	}
	c, err := etcdCli.New(cfg)
	if err != nil {
		return e.Forward(err)
	}
	etc.kapi = etcdCli.NewKeysAPI(c)

	if etc.SecKeyRing == "" {
		return nil
	}

	kr, err := os.Open(etc.SecKeyRing)
	if err != nil {
		return e.Forward(err)
	}
	defer kr.Close()
	etc.cm, err = config.NewEtcdConfigManager(etc.Endpoints, kr)
	if err != nil {
		return e.Forward(err)
	}

	etc.keystore, err = ioutil.ReadFile(etc.SecKeyRing)
	if err != nil {
		return e.Forward(err)
	}

	return nil
}

func (etc *Etcd) Put(key string, buf []byte) error {
	var err error
	if etc.SecKeyRing == "" {
		_, err = etc.kapi.Create(context.Background(), key, string(buf))
		if e.Contains(err, "Key already exists") {
			_, err = etc.kapi.Update(context.Background(), key, string(buf))
			if err != nil {
				return e.Forward(err)
			}
		} else if err != nil {
			return e.Forward(err)
		}
		return nil
	}
	err = etc.cm.Set(key, buf)
	if err != nil {
		return e.Forward(err)
	}
	return nil
}

func (etc *Etcd) Get(key string, opt *etcdCli.GetOptions) ([]byte, error) {
	var err error
	if etc.SecKeyRing == "" {
		if opt == nil {
			opt = &etcdCli.GetOptions{}
		}
		resp, err := etc.kapi.Get(context.Background(), key, opt)
		if err != nil {
			return nil, e.Forward(err)
		}
		if len(resp.Node.Nodes) > 0 {
			fmt.Println(resp.Node.Nodes)
		}
		return []byte(resp.Node.String()), nil
	}
	buf, err := etc.cm.Get(key)
	if err != nil {
		return nil, e.Forward(err)
	}
	return buf, nil
}

func (etc *Etcd) GetNodes(key string, opt *etcdCli.GetOptions) (etcdCli.Nodes, error) {
	if etc.SecKeyRing == "" {
		if opt == nil {
			opt = &etcdCli.GetOptions{}
		}
		resp, err := etc.kapi.Get(context.Background(), key, opt)
		if err != nil {
			return nil, e.Forward(err)
		}
		if len(resp.Node.Nodes) == 0 {
			return nil, e.New("empty nodes")
		}
		return resp.Node.Nodes, nil
	}
	return nil, e.New("no support for crypt etcdCli")
}

func (etc *Etcd) TryGetNodes(key string, opt *etcdCli.GetOptions, timeout time.Duration) (nodes etcdCli.Nodes, err error) {
	op := func() (err error) {
		nodes, err = etc.GetNodes(key, opt)
		if err != nil {
			err = e.Forward(err)
			return
		}
		return
	}
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = timeout
	err = backoff.Retry(op, backOff)
	if err != nil {
		err = e.Forward(err)
		return
	}
	return
}

func (etc *Etcd) TryGetUntil(key string, timeout time.Duration) (buf []byte, err error) {
	op := func() (err error) {
		buf, err = etc.Get(key, nil)
		if err != nil {
			err = e.Forward(err)
			return
		}
		return
	}
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = timeout
	err = backoff.Retry(op, backOff)
	if err != nil {
		err = e.Forward(err)
		return
	}
	return
}

func (etc *Etcd) TryGetRespUntil(key string, timeout time.Duration) (resp *etcdCli.Response, err error) {
	op := func() (err error) {
		resp, err = etc.kapi.Get(context.Background(), key, &etcdCli.GetOptions{})
		if err != nil {
			return
		}
		return
	}
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = timeout
	err = backoff.Retry(op, backOff)
	if err != nil {
		err = e.Forward(err)
		return
	}
	if etc.SecKeyRing == "" {
		return
	}
	if resp.Node != nil && len(etc.keystore) > 0 {
		var buf []byte
		buf, err = secconf.Decode([]byte(resp.Node.Value), bytes.NewBuffer(etc.keystore))
		if err != nil {
			err = e.Forward(err)
			return
		}
		resp.Node.Value = string(buf)
	}
	return
}

func (etc *Etcd) Del(key string, opt *etcdCli.DeleteOptions) error {
	_, err := etc.kapi.Delete(context.Background(), key, opt)
	if err != nil {
		return e.Forward(err)
	}
	return nil
}

type watcher struct {
	w        etcdCli.Watcher
	keystore []byte
}

func (w watcher) Next(ctx context.Context) (*etcdCli.Response, error) {
	resp, err := w.w.Next(ctx)
	if err != nil {
		return nil, err
	}
	if resp.Node != nil && len(w.keystore) > 0 {
		buf, err := secconf.Decode([]byte(resp.Node.Value), bytes.NewBuffer(w.keystore))
		if err != nil {
			return nil, e.Forward(err)
		}
		resp.Node.Value = string(buf)
	}
	// if resp.PrevNode != nil {
	// 	resp.PrevNode.Value = secconf.Decode(resp.PrevNode.Value, bytes.NewBuffer(c.keystore))
	// }
	return resp, nil
}

func (etc *Etcd) Watcher(key string, opts *etcdCli.WatcherOptions) (etcdCli.Watcher, error) {
	w := etc.kapi.Watcher(key, opts)
	return &watcher{
		w:        w,
		keystore: etc.keystore,
	}, nil
}

// type Instance struct {
// 	Name     string
// 	Session  string
// 	Instance string
// 	Pid      string
// 	Proto    string
// 	Addr     string
// 	Hostname string
// }
//
// func (i *Instance) Serialize() ([]byte, error) {
// 	buf, err := yaml.Marshal(i)
// 	if err != nil {
// 		return nil, e.Forward(err)
// 	}
// 	return buf, nil
// }
//
// func NewInstance(in []byte) (*Instance, error) {
// 	inst := new(Instance)
// 	err := yaml.Unmarshal(in, inst)
// 	if err != nil {
// 		return nil, e.Forward(err)
// 	}
// 	return inst, nil
// }
//
// func RegisterService(etc *Etcd, name, sess, inst string, server *core.Server) {
// 	hostname, _ := os.Hostname()
// 	i := &Instance{
// 		Name:     name,
// 		Session:  sess,
// 		Instance: inst,
// 		Pid:      strconv.FormatInt(int64(os.Getpid()), 10),
// 		Proto:    server.Proto,
// 		Addr:     server.Address().String(),
// 		Hostname: hostname,
// 	}
// 	buf, err := i.Serialize()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	err = etc.Put(filepath.Join(Name, "gormethods", "instances", name), buf)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Printf("Service %v registered.", name)
// }
//
// func RegisterServiceReadInst(etc *Etcd, name string, etcdTimeout time.Duration) (inst, proto, addr string, err error) {
// 	buf, err := etc.TryGetUntil(
// 		filepath.Join(Name, "gormethods", "instances", name),
// 		etcdTimeout,
// 	)
// 	if err != nil {
// 		err = e.Forward(err)
// 		return
// 	}
// 	i, err := NewInstance([]byte(buf))
// 	if err != nil {
// 		err = e.Forward(err)
// 		return
// 	}
// 	inst = i.Instance
// 	proto = i.Proto
// 	addr = i.Addr
// 	return
// }
