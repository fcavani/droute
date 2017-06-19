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
	"github.com/fcavani/log"
)

// DefaultLang is the default language. Need to be setted in the starp of the
// application.
var DefaultLang = "en"

// RequestConfig is the configuration for the request parse and pre processing.
type RequestConfig struct {
	AcceptedUserAgents *useragent.UAVers
	AcceptedLanguages  []string
	AcceptedEncoding   []string
	AcceptedMedias     []string
	AcceptedCharsets   []string
}

// DefaultRequestConfig is a simple default config for request.
var DefaultRequestConfig *RequestConfig

func init() {
	rc := &RequestConfig{
		AcceptedUserAgents: useragent.NewUAVers(),
		AcceptedLanguages:  []string{"pt-br", "en"},
		AcceptedEncoding:   []string{"gzip", "compress", "deflate", "identity"},
		AcceptedMedias:     []string{"text/html", "image/jpeg", "image/png", "image/webp", "image/gif"},
		AcceptedCharsets:   []string{"utf-8"},
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
	DefaultRequestConfig = rc
}

// UserRequest provide methods to access the user request information.
type UserRequest struct {
	rc       *RequestConfig
	r        *http.Request
	language string
	ua       *useragent.Product
	ip       string
	url      *url.URL
	referrer *url.URL
	query    map[string]string
	charset  string
	media    string
	encoding string
	timezone *time.Location
}

func (ur *UserRequest) Media() string {
	if ur.media != "" {
		return ur.media
	}
	ur.media = getparam(ur.r, "Accept", ur.rc.AcceptedMedias, typeparams.Parse)
	return ur.media
}

func (ur *UserRequest) Charset() string {
	if ur.charset != "" {
		return ur.charset
	}
	ur.charset = getparam(ur.r, "Accept-Charset", ur.rc.AcceptedCharsets, typeparams.Parse)
	return ur.charset
}

func (ur *UserRequest) Language() string {
	fromReq := getparam(ur.r, "Accept-Language", ur.rc.AcceptedLanguages, typeparams.ParseLang)
	fromUrl := ""
	if query := ur.Query(); query != nil && query["lang"] != "" {
		err := typeparams.CheckTypeParams(query["lang"], 2, 6)
		if err == nil {
			fromUrl = query["lang"]
		}
	}
	// fromD := ""
	// if ur.D != nil {
	// 	if ilang, err := ur.D.Get("lang"); err == nil && ilang != nil {
	// 		fromD = ilang.(string)
	// 	}
	// }
	// 0o. Manter o estado anterios se fromUrl não mudar
	// 1o. fromUrl
	// 2o. fromReq
	// 3o. ur.language
	// 4o. default
	// defer func() {
	// 	if ur.D != nil {
	// 		ur.D.Set("lang", ur.language)
	// 	}
	// }()
	if fromUrl == "" && ur.language != "" {
		return ur.language
		// } else if fromUrl == "" && fromD != "" {
		// 	ur.language = fromD
	} else if fromUrl != "" {
		ur.language = fromUrl
	} else if fromReq != "" {
		ur.language = fromReq
	} else if ur.language != "" {
		return ur.language
	} else {
		ur.language = DefaultLang
	}

	return ur.language
}

func (ur *UserRequest) Encoding() string {
	if ur.encoding != "" {
		return ur.encoding
	}
	ur.encoding = getparam(ur.r, "Accept-Encoding", ur.rc.AcceptedEncoding, typeparams.Parse)
	return ur.encoding
}

func (ur *UserRequest) LangDir() string {
	return "ltr"
}

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

func (ur *UserRequest) Ip() string {
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

func (ur *UserRequest) Url() *url.URL {
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

func (ur *UserRequest) Query() map[string]string {
	if ur.query != nil {
		return ur.query
	}
	var err error
	ur.query, err = h.GetUrlQuery(ur.Url().RawQuery)
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
		return time.FixedZone("UTC±0", 0), nil
	}
	utc += strconv.Itoa(h)
	m := int((float32(minutes)/60.0 - float32(h)) * 60.0)
	if m > 0 {
		utc += ":" + strconv.Itoa(m)
	}
	return time.FixedZone(utc, minutes*-60), nil
}

func (ur *UserRequest) Timezone() *time.Location {
	var tzstr string
	if query := ur.Query(); query != nil && query["tz"] != "" {
		err := typeparams.CheckTypeParams(query["tz"], 2, 6)
		if err == nil {
			tzstr = query["tz"]
		}
	}
	if tzstr == "" {
		if ur.timezone != nil {
			return ur.timezone
		}
		return time.Local
	}
	loc, err := timeZoneJS2utc(tzstr)
	if err != nil {
		log.Tag("request").Errorf("Can't convert time zone: %v", err)
		return nil
	}
	return loc
}

// Request get the user request from the context.
func Request(r *http.Request) *UserRequest {
	return r.Context().Value("request").(*UserRequest)
}

// RequestHandler reads the information in the request parameter and process it,
// the result got to the context under the name "request". All data from the
// request must be accessed with the function Request(r), i.e. Request(r).Referrer().
// This is done with the intent of make the information in r more acessible and
// in this way avaid repetitive work.
func RequestHandler(rc *RequestConfig, formsizefile int64, handler http.HandlerFunc) http.HandlerFunc {
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
		param = accepted[0]
		return
	}
	params = strings.ToLower(params)
	parsed, err := parser(params)
	if err != nil {
		param = accepted[0]
		return
	}
	param = parsed.FindBest(accepted)
	if param == "" {
		param = accepted[0]
	}
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
		log.Tag("router", "server", "proxy").Error(er)
	}
}
