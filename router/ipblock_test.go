package router

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/fcavani/droute/list"
	"github.com/fcavani/droute/responsewriter"
	h "github.com/fcavani/http"
	"github.com/spf13/viper"
)

func TestIPBlockDeny(t *testing.T) {
	viper.SetConfigType("yaml")
	cfg := "denny:\n - 127.0.0.1"
	buf := bytes.NewBufferString(cfg)
	err := viper.ReadConfig(buf)
	if err != nil {
		t.Fatal(err)
	}
	dennyList, err := list.NewRegexpList("denny")
	if err != nil {
		t.Fatal(err)
	}
	h := IPBlock(dennyList, func(rw *responsewriter.ResponseWriter, r *http.Request) {
		ip, _ := h.RemoteIP(r)
		t.Log("IP:", ip)
		rw.WriteHeader(200)
	})
	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("X-Real-Ip", "127.0.0.1")
	rw := responsewriter.NewResponseWriter()
	h(rw, r)
	if code := rw.ResponseCode(); code != 403 {
		t.Fatal("wrong response code", code)
	}
}

func TestIPBlockAllow(t *testing.T) {
	viper.SetConfigType("yaml")
	cfg := "denny:\n - 192.168.1.1"
	buf := bytes.NewBufferString(cfg)
	err := viper.ReadConfig(buf)
	if err != nil {
		t.Fatal(err)
	}
	dennyList, err := list.NewRegexpList("denny")
	if err != nil {
		t.Fatal(err)
	}
	h := IPBlock(dennyList, func(rw *responsewriter.ResponseWriter, r *http.Request) {
		ip, _ := h.RemoteIP(r)
		t.Log("IP:", ip)
		rw.WriteHeader(200)
	})
	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("X-Real-Ip", "10.0.0.1")
	rw := responsewriter.NewResponseWriter()
	h(rw, r)
	if code := rw.ResponseCode(); code != 200 {
		t.Fatal("wrong response code", code)
	}
}
