// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package request

import (
	"net/http"
	"testing"
	"time"

	"github.com/fcavani/droute/responsewriter"
	"github.com/fcavani/e"
)

func TestMedia(t *testing.T) {
	ur, err := newUserRequest("Accept", "text/html")
	if err != nil {
		t.Fatal(err)
	}
	if m := ur.Media(); m != "text/html" {
		t.Fatal("invalid media", m)
	}
	if m := ur.Media(); m != "text/html" {
		t.Fatal("invalid media", m)
	}
	ur, err = newUserRequest("Accept", "text/foo")
	if err != nil {
		t.Fatal(err)
	}
	if m := ur.Media(); m != "" {
		t.Fatal("invalid media", m)
	}
	ur, err = newUserRequest("Accept", "image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	if m := ur.Media(); m != "image/jpeg" {
		t.Fatal("invalid media", m)
	}
	ur, err = newUserRequest("", "")
	if err != nil {
		t.Fatal(err)
	}
	if m := ur.Media(); m != "" {
		t.Fatal("invalid media", m)
	}
}

func TestCharset(t *testing.T) {
	ur, err := newUserRequest("Accept-Charset", "utf-8")
	if err != nil {
		t.Fatal(err)
	}
	if m := ur.Charset(); m != "utf-8" {
		t.Fatal("invalid charset", m)
	}
	if m := ur.Charset(); m != "utf-8" {
		t.Fatal("invalid charset", m)
	}
}

func TestLanguage(t *testing.T) {
	ur, err := newUserRequest("Accept-Language", "pt-br")
	if err != nil {
		t.Fatal(err)
	}
	if m := ur.Language(); m != "pt-br" {
		t.Fatal("invalid lang", m)
	}
	if m := ur.Language(); m != "pt-br" {
		t.Fatal("invalid lang", m)
	}
	r, err := http.NewRequest("GET", "http://localhost?lang=en", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("Accept-Language", "pt-br")
	ur = &UserRequest{
		rc: DefaultConfig,
		r:  r,
	}
	if m := ur.Language(); m != "en" {
		t.Fatal("invalid lang", m)
	}
	if m := ur.Language(); m != "en" {
		t.Fatal("invalid lang", m)
	}
	ur, err = newUserRequest("", "")
	if err != nil {
		t.Fatal(err)
	}
	if m := ur.Language(); m != "en" {
		t.Fatal("invalid lang", m)
	}
}

func TestEncoding(t *testing.T) {
	ur, err := newUserRequest("Accept-Encoding", "gzip")
	if err != nil {
		t.Fatal(err)
	}
	if m := ur.Encoding(); m != "gzip" {
		t.Fatal("invalid encoding", m)
	}
	if m := ur.Encoding(); m != "gzip" {
		t.Fatal("invalid encoding", m)
	}
}

func TestLangDir(t *testing.T) {
	ur, err := newUserRequest("", "")
	if err != nil {
		t.Fatal(err)
	}
	if m := ur.LangDir(); m != "ltr" {
		t.Fatal("invalid lang dir", m)
	}
}

func TestUseragent(t *testing.T) {
	ur, err := newUserRequest("User-Agent", "Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US) AppleWebKit/534.3 (KHTML, like Gecko) Chrome/6.0.464.0 Safari/534.3")
	if err != nil {
		t.Fatal(err)
	}
	if p := ur.Useragent(); p.Name != "Chrome" && p.Version.String() != "6.0.464.0" {
		t.Fatal("invalid ua", p)
	}
	if p := ur.Useragent(); p.Name != "Chrome" && p.Version.String() != "6.0.464.0" {
		t.Fatal("invalid ua", p)
	}
	ur, err = newUserRequest("", "")
	if err != nil {
		t.Fatal(err)
	}
	if p := ur.Useragent(); p.Name != "Mozilla" && p.Version.String() != "4" {
		t.Fatal("invalid ua", p)
	}
	if p := ur.Useragent(); p.Name != "Mozilla" && p.Version.String() != "4" {
		t.Fatal("invalid ua", p)
	}
	ur, err = newUserRequest("User-Agent", "catoto/2.0")
	if err != nil {
		t.Fatal(err)
	}
	if p := ur.Useragent(); p.Name != "Mozilla" && p.Version.String() != "4" {
		t.Fatal("invalid ua", p)
	}
}

func TestIP(t *testing.T) {
	ur, err := newUserRequest("X-Real-Ip", "10.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if m := ur.IP(); m != "10.0.0.1" {
		t.Fatal("invalid ip", m)
	}
	if m := ur.IP(); m != "10.0.0.1" {
		t.Fatal("invalid ip", m)
	}
}

func TestURL(t *testing.T) {
	r, err := http.NewRequest("GET", "http://localhost/dir/index.html?q=teste&u=1#frag", nil)
	if err != nil {
		t.Fatal(err)
	}
	ur := &UserRequest{
		rc: DefaultConfig,
		r:  r,
	}
	u := ur.URL()
	if host := u.Hostname(); host != "localhost" {
		t.Fatal("invalid url", host)
	}
	if u.Path != "/dir/index.html" {
		t.Fatal("invalid url", u.Path)
	}
	q := ur.Query()
	if len(q) != 2 {
		t.Fatal("invalid query")
	}
	if n, found := q["q"]; !found || n != "teste" {
		t.Fatal("invalid query")
	}
	if n, found := q["u"]; !found || n != "1" {
		t.Fatal("invalid query")
	}
	if u.Fragment != "frag" {
		t.Fatal("invalid fragment", u.Fragment)
	}
}

func TestReferrer(t *testing.T) {
	ur, err := newUserRequest("Referer", "http://nowareplace")
	if err != nil {
		t.Fatal(err)
	}
	if m := ur.Referrer(); m.String() != "http://nowareplace" {
		t.Fatal("invalid referrer", m)
	}
	if m := ur.Referrer(); m.String() != "http://nowareplace" {
		t.Fatal("invalid referrer", m)
	}
}

func TestTimeZoneJS2utc(t *testing.T) {
	l, err := timeZoneJS2utc("boo")
	if err != nil && !e.Contains(err, "invalid syntax") {
		t.Fatal(err)
	}
	if l != time.Local {
		t.Fatal("invalid location", l)
	}
	l, err = timeZoneJS2utc("0")
	if err != nil {
		t.Fatal(err)
	}
	if l != time.UTC {
		t.Fatal("invalid location", l)
	}
	l, err = timeZoneJS2utc("182")
	if err != nil {
		t.Fatal(err)
	}
	if l.String() != "UTC-3:1" {
		t.Fatal("invalid location", l)
	}
}

func TestTimezone(t *testing.T) {
	// UTC-03
	r, err := http.NewRequest("GET", "http://localhost/?tz=180", nil)
	if err != nil {
		t.Fatal(err)
	}
	ur := &UserRequest{
		rc: DefaultConfig,
		r:  r,
	}
	if m := ur.Timezone(); m.String() != "UTC-3" {
		t.Fatal("invalid tz", m.String())
	}
	if m := ur.Timezone(); m.String() != "UTC-3" {
		t.Fatal("invalid tz", m.String())
	}

	r, err = http.NewRequest("GET", "http://localhost/?tz=-180", nil)
	if err != nil {
		t.Fatal(err)
	}
	ur = &UserRequest{
		rc: DefaultConfig,
		r:  r,
	}
	if m := ur.Timezone(); m.String() != "UTC+3" {
		t.Fatal("invalid tz", m.String())
	}

	r, err = http.NewRequest("GET", "http://localhost/?tz=bazinga", nil)
	if err != nil {
		t.Fatal(err)
	}
	ur = &UserRequest{
		rc: DefaultConfig,
		r:  r,
	}
	if m := ur.Timezone(); m != time.Local {
		t.Fatal("invalid tz", m)
	}

	ur, err = newUserRequest("", "")
	if err != nil {
		t.Fatal(err)
	}
	if m := ur.Timezone(); m != time.Local {
		t.Fatal("invalid tz", m)
	}
}

func TestHandler(t *testing.T) {
	h := Handler(DefaultConfig, 10240, func(w http.ResponseWriter, r *http.Request) {
		ur := Request(r)
		if c := ur.Charset(); c != "utf-8" {
			t.Fatal("invalid charset", c)
		}
		w.WriteHeader(200)
	})
	w := responsewriter.NewResponseWriter()
	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("Accept-Charset", "utf-8")
	h(w, r)
	if code := w.ResponseCode(); code != 200 {
		t.Fatal("handler failed", code)
	}
}

func newUserRequest(key, value string) (*UserRequest, error) {
	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		return nil, err
	}
	if key != "" && value != "" {
		r.Header.Add(key, value)
	}
	ur := &UserRequest{
		rc: DefaultConfig,
		r:  r,
	}
	return ur, nil
}
