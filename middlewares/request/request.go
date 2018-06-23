// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package request

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fcavani/e"
	h "github.com/fcavani/http"
	"github.com/fcavani/http/typeparams"
	"github.com/fcavani/http/useragent"
	"github.com/fcavani/net/dns"
	log "github.com/fcavani/slog"
)

// DefaultLang is the default language. Need to be setted in the starup of the
// application.
var DefaultLang = "en"

// Config is the configuration for the request parse and pre processing.
type Config struct {
	AcceptedUserAgents *useragent.UAVers
	AcceptedLanguages  []string
	AcceptedEncoding   []string
	AcceptedMedias     []string
	AcceptedCharsets   []string
	Localhosts         []string
}

// DefaultRequestConfig is a simple default config for request.
var DefaultConfig *Config

func init() {
	rc := &Config{
		AcceptedUserAgents: useragent.NewUAVers(),
		AcceptedLanguages:  []string{"pt-br", "en"},
		AcceptedEncoding:   []string{"gzip", "compress", "deflate", "identity"},
		AcceptedMedias:     []string{"text/html", "image/jpeg", "image/png", "image/webp", "image/gif"},
		AcceptedCharsets:   []string{"utf-8"},
		Localhosts:         []string{},
	}

	err := rc.AcceptedUserAgents.Push("Mozilla", ">=4")
	if err != nil {
		log.Fatal(err)
	}
	err = rc.AcceptedUserAgents.Push("Safari", ">=600")
	if err != nil {
		log.Fatal(err)
	}
	err = rc.AcceptedUserAgents.Push("Firefox", ">=4")
	if err != nil {
		log.Fatal(err)
	}
	err = rc.AcceptedUserAgents.Push("Chrome", ">=6")
	if err != nil {
		log.Fatal(err)
	}
	DefaultConfig = rc
}

// UserRequest provide methods to access the user request information.
type UserRequest struct {
	rc              *Config
	r               *http.Request
	language        string
	ua              *useragent.Product
	ip              string
	url             *url.URL
	referrer        *url.URL
	query           map[string]string
	charset         string
	media           string
	encoding        string
	timezone        *time.Location
	localhosts      bool
	localhostCached bool
}

// IsLocalhost is local host.
func (ur *UserRequest) IsLocalhost() bool {
	if ur.localhostCached {
		return ur.localhosts
	}
	for _, host := range ur.rc.Localhosts {
		addrs, err := dns.LookupHost(host)
		if err != nil {
			return false
		}
		for _, addr := range addrs {
			if addr == ur.IP() {
				ur.localhostCached = true
				ur.localhosts = true
				return true
			}
		}
	}
	ur.localhostCached = true
	ur.localhosts = false
	return false
}

// Media return the accepted media by the user.
func (ur *UserRequest) Media() string {
	if ur.media != "" {
		return ur.media
	}
	ur.media = getparam(ur.r, "Accept", ur.rc.AcceptedMedias, typeparams.Parse)
	return ur.media
}

// Charset return the accepted charset by the user.
func (ur *UserRequest) Charset() string {
	if ur.charset != "" {
		return ur.charset
	}
	ur.charset = getparam(ur.r, "Accept-Charset", ur.rc.AcceptedCharsets, typeparams.Parse)
	return ur.charset
}

// Language return the language accepted by the user.
func (ur *UserRequest) Language() string {
	fromReq := getparam(ur.r, "Accept-Language", ur.rc.AcceptedLanguages, typeparams.ParseLang)
	fromURL := ""
	if query := ur.Query(); query != nil && query["lang"] != "" {
		err := typeparams.CheckTypeParams(query["lang"], 2, 6)
		if err == nil {
			fromURL = query["lang"]
		}
	}
	// fromD := ""
	// if ur.D != nil {
	// 	if ilang, err := ur.D.Get("lang"); err == nil && ilang != nil {
	// 		fromD = ilang.(string)
	// 	}
	// }
	// 0o. Manter o estado anterios se fromURL não mudar
	// 1o. fromURL
	// 2o. fromReq
	// 3o. ur.language
	// 4o. default
	// defer func() {
	// 	if ur.D != nil {
	// 		ur.D.Set("lang", ur.language)
	// 	}
	// }()
	if fromURL == "" && ur.language != "" {
		return ur.language
		// } else if fromURL == "" && fromD != "" {
		// 	ur.language = fromD
	} else if fromURL != "" {
		ur.language = fromURL
	} else if fromReq != "" {
		ur.language = fromReq
	} else if ur.language != "" {
		return ur.language
	} else {
		ur.language = DefaultLang
	}

	return ur.language
}

// Encoding return the encoding algorithm accepted by the user.
func (ur *UserRequest) Encoding() string {
	if ur.encoding != "" {
		return ur.encoding
	}
	ur.encoding = getparam(ur.r, "Accept-Encoding", ur.rc.AcceptedEncoding, typeparams.Parse)
	return ur.encoding
}

// LangDir is the language write direction, always return ltr.
func (ur *UserRequest) LangDir() string {
	return "ltr"
}

// Useragent return the user ua.
func (ur *UserRequest) Useragent() *useragent.Product {
	if ur.ua != nil {
		return ur.ua
	}
	var err error
	uastr := ur.r.UserAgent()
	if uastr == "" {
		ur.ua, err = ur.rc.AcceptedUserAgents.Best()
		if err != nil {
			return nil
		}
	} else {
		uav := useragent.NewUserAgentDetector()
		ur.ua, err = uav.FindMatch(uastr, ur.rc.AcceptedUserAgents)
		if err != nil {
			ur.ua, err = ur.rc.AcceptedUserAgents.Best()
			if err != nil {
				return nil
			}
		}
	}
	return ur.ua
}

// IP returns the users real ip.
func (ur *UserRequest) IP() string {
	if ur.ip != "" {
		return ur.ip
	}
	var err error
	ur.ip, err = h.RemoteIP(ur.r)
	if err != nil {
		return ""
	}
	return ur.ip
}

// URL return the user requested url.
func (ur *UserRequest) URL() *url.URL {
	if ur.url != nil {
		return ur.url
	}
	var err error
	// root := ur.Root
	// if root == "/" {
	// 	root = ""
	// }
	root := ""
	ur.url, err = h.Url(ur.r, root)
	if err != nil {
		return nil
	}
	return ur.url
}

// Referrer return the referrer url.
func (ur *UserRequest) Referrer() *url.URL {
	if ur.referrer != nil {
		return ur.referrer
	}
	var err error
	ur.referrer, err = url.Parse(ur.r.Referer())
	if err != nil {
		return nil
	}
	return ur.referrer
}

// Query returns a map with the queries in the url made by the user.
func (ur *UserRequest) Query() map[string]string {
	if ur.query != nil {
		return ur.query
	}
	var err error
	ur.query, err = h.GetUrlQuery(ur.URL().RawQuery)
	if err != nil {
		return nil
	}
	return ur.query
}

func timeZoneJS2utc(tzstr string) (*time.Location, error) {
	//js: var offset = new Date().getTimezoneOffset();
	//The time-zone offset is the difference, in minutes, between UTC and local
	//time. Note that this means that the offset is positive if the local
	//timezone is behind UTC and negative if it is ahead. For example, if your
	//time zone is UTC+10 (Australian Eastern Standard Time), -600 will be
	//returned. Daylight savings time prevents this value from being a constant
	//even for a given locale
	minutes, err := strconv.Atoi(tzstr)
	if err != nil {
		return time.Local, e.Forward(err)
	}
	utc := "UTC-"
	if minutes < 0 {
		utc = "UTC+"
		minutes *= -1
	}
	h := minutes / 60
	if h == 0 {
		//return time.FixedZone("UTC±0", 0), nil
		return time.UTC, nil
	}
	utc += strconv.Itoa(h)
	m := int((float32(minutes)/60.0 - float32(h)) * 60.0)
	if m > 0 {
		utc += ":" + strconv.Itoa(m)
	}
	return time.FixedZone(utc, minutes*-60), nil
}

//Timezone receives the time zone from que query parameter tz in the javascript
// format and convert it to time.Location.
func (ur *UserRequest) Timezone() *time.Location {
	var err error
	var tzstr string
	if ur.timezone != nil {
		return ur.timezone
	}
	if query := ur.Query(); query != nil && query["tz"] != "" {
		err = typeparams.CheckTypeParams(query["tz"], 2, 6)
		if err == nil {
			tzstr = query["tz"]
		}
	}
	if tzstr == "" {
		ur.timezone = time.Local
		return ur.timezone
	}
	ur.timezone, err = timeZoneJS2utc(tzstr)
	if err != nil {
		log.Tag("request").Errorf("Can't convert time zone: %v", err)
		return nil
	}
	return ur.timezone
}

// Request get the user request from the context.
func Request(r *http.Request) *UserRequest {
	return r.Context().Value("request").(*UserRequest)
}

// Handler reads the information in the request parameter and process it,
// the result got to the context under the name "request". All data from the
// request must be accessed with the function Request(r), i.e. Request(r).Referrer().
// This is done with the intent of make the information in r more acessible and
// in this way avaid repetitive work.
func Handler(rc *Config, formsizefile int64, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			err = e.Push(err, "Can't parse the form.")
			errHandler(w, http.StatusInternalServerError, err)
			return
		}

		err = r.ParseMultipartForm(formsizefile)
		if err != nil && !e.Contains(err, "request Content-Type isn't multipart/form-data") {
			err = e.Push(err, "Can't parse the form.")
			errHandler(w, http.StatusInternalServerError, err)
			return
		}
		req := &UserRequest{
			rc: rc,
			r:  r,
		}
		r = r.WithContext(context.WithValue(r.Context(), "request", req))
		handler(w, r)
	}
}

func getparam(r *http.Request, header string, accepted []string, parser func(string) (typeparams.TypeParams, error)) (param string) {
	params := r.Header.Get(header)
	if params == "" {
		return
	}
	params = strings.ToLower(params)
	parsed, err := parser(params)
	if err != nil {
		return
	}
	param = parsed.FindBest(accepted)
	return
}

type msgErr struct {
	Err string
}

func errHandler(w http.ResponseWriter, code int, err error) {
	if err == nil {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	resp := msgErr{
		Err: err.Error(),
	}
	er := json.NewEncoder(w).Encode(resp)
	if er != nil {
		log.Tag("router", "server").Error(er)
	}
}
