// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package errhandler

import (
	"encoding/json"
	"net/http"

	"github.com/fcavani/e"
	log "github.com/fcavani/slog"
)

// ErrHandler trow a json error message with the code and the error.
func ErrHandler(w http.ResponseWriter, code int, err error) {
	if err == nil {
		return
	}
	log.Tag("error", "handlers").DebugLevel().Println(e.Trace(e.Forward(err)))
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	resp := struct {
		Code  int    `jsong:"code"`
		Err   string `json:"err"`
		Human string `json:"human"`
	}{
		Code:  code,
		Err:   err.Error(),
		Human: e.Human(err),
	}
	er := json.NewEncoder(w).Encode(resp)
	if er != nil {
		log.Tag("router", "server", "proxy").Error(er)
	}
}
