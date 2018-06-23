package fileserver

import (
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/fcavani/httprouter.v2"
)

func StripPrefix(prefix string, h Handler) http.Handler {
	if prefix == "" {
		return h
	}
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		lang := httprouter.ContentLang(req)
		if lang != "" {
			p := strings.TrimPrefix(path, "/"+lang)
			if len(p) < len(path) {
				path = p
			}
		}
		if p := strings.TrimPrefix(path, prefix); len(p) < len(path) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p
			h.ServeHTTP(w, r2)
		} else {
			http.NotFound(w, r)
		}
	})
}
