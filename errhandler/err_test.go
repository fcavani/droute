// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package errhandler

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/fcavani/droute/responsewriter"
)

func TestErrHandler(t *testing.T) {
	rw := responsewriter.NewResponseWriter()
	ErrHandler(rw, 200, errors.New("this is a error"))
	if rw.ResponseCode() != 200 {
		t.Fatal("wrong error code")
	}
	buf := rw.Bytes()
	if buf == nil || len(buf) == 0 {
		t.Fatal("something wrong with the response buffer")
	}
	var resp struct {
		Err string `json:"err"`
	}
	err := json.Unmarshal(buf, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != "this is a error" {
		t.Fatal("wrong error")
	}
}
