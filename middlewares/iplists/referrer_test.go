// Copyright © 2017 Felipe A. Cavani <fcavani@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iplists

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/fcavani/droute/list"
	"github.com/fcavani/droute/middlewares/request"
	"github.com/fcavani/droute/responsewriter"
	"github.com/spf13/viper"
)

func TestReferrerDeny(t *testing.T) {
	h := request.Handler(request.DefaultConfig, 1024, Referrer(DENY, nil, nil, nil,
		func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(200)
		}),
	)
	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	rw := responsewriter.NewResponseWriter()
	h(rw, r)
	if code := rw.ResponseCode(); code != 403 {
		t.Fatal("wrong response code", code)
	}
}

func TestReferrerAllow(t *testing.T) {
	h := request.Handler(request.DefaultConfig, 1024, Referrer(ALLOW, nil, nil, nil,
		func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(200)
		}),
	)
	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	rw := responsewriter.NewResponseWriter()
	h(rw, r)
	if code := rw.ResponseCode(); code != 200 {
		t.Fatal("wrong response code", code)
	}
}

func TestReferrerDenyList(t *testing.T) {
	viper.SetConfigType("yaml")
	cfg := "denny:\n - '^([^.]+.)*?hell.com'"
	buf := bytes.NewBufferString(cfg)
	err := viper.ReadConfig(buf)
	if err != nil {
		t.Fatal(err)
	}
	dennyList, err := list.NewRegexpList("denny")
	if err != nil {
		t.Fatal(err)
	}
	h := request.Handler(request.DefaultConfig, 1024, Referrer(ALLOW, nil, dennyList, nil,
		func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(200)
		}),
	)
	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("Referer", "http://hell.com")
	rw := responsewriter.NewResponseWriter()
	h(rw, r)
	if code := rw.ResponseCode(); code != 403 {
		t.Fatal("wrong response code", code)
	}
}

func TestReferrerAllowList(t *testing.T) {
	viper.SetConfigType("yaml")
	cfg := "allow:\n - '^([^.]+.)*?heaven.com'"
	buf := bytes.NewBufferString(cfg)
	err := viper.ReadConfig(buf)
	if err != nil {
		t.Fatal(err)
	}
	allowList, err := list.NewRegexpList("allow")
	if err != nil {
		t.Fatal(err)
	}
	h := request.Handler(request.DefaultConfig, 1024, Referrer(DENY, allowList, nil, nil,
		func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(200)
		}),
	)
	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("Referer", "http://heaven.com")
	rw := responsewriter.NewResponseWriter()
	h(rw, r)
	if code := rw.ResponseCode(); code != 200 {
		t.Fatal("wrong response code", code)
	}
}
