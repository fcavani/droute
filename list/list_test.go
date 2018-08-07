// Copyright © 2018 Felipe A. Cavani <fcavani@gmail.com>
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

package list

import (
	"bytes"
	"testing"

	"github.com/fcavani/e"
	"github.com/spf13/viper"
)

func TestRegexp(t *testing.T) {
	viper.SetConfigType("yaml")
	cfg := "denny:\n - '^([^.]+.)*?buttons\\.com'\nlist:\n - 'text'"
	buf := bytes.NewBufferString(cfg)
	err := viper.ReadConfig(buf)
	if err != nil {
		t.Fatal(err)
	}

	re, err := NewRegexpList("denny")
	if err != nil {
		t.Fatal(err)
	}

	if !re.Exist("buttons.com") {
		t.Fatal("fail")
	}

	if re.Exist("foo.com") {
		t.Fatal("fail")
	}
}

func TestRegexpUpdate(t *testing.T) {
	re, err := NewRegexpList("denny")
	if err != nil {
		t.Fatal(err)
	}

	cfg := "denny:\n - '^([^.]+.)*?referrer\\.com'\nlist:\n - 'txt'"
	buf := bytes.NewBufferString(cfg)
	err = viper.ReadConfig(buf)
	if err != nil {
		t.Fatal(err)
	}

	re.Update("denny")

	if !re.Exist("referrer.com") {
		t.Fatal("fail")
	}

	if re.Exist("foo.com") {
		t.Fatal("fail")
	}
}

func TestRegexpNotList(t *testing.T) {
	_, err := NewRegexpList("nolist")
	if err != nil && !e.Contains(err, "can't read the list from viper config") {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error is nil")
	}
}

func TestTextList(t *testing.T) {
	tlist, err := NewTextList("list")
	if err != nil {
		t.Fatal(err)
	}

	if !tlist.Exist("txt") {
		t.Fatal("fail")
	}

	if tlist.Exist("bar") {
		t.Fatal("fail")
	}
}

func TestTextListUpdate(t *testing.T) {
	tlist, err := NewTextList("list")
	if err != nil {
		t.Fatal(err)
	}

	cfg := "denny:\n - '^([^.]+.)*?referrer\\.com'\nlist:\n - 'text'"
	buf := bytes.NewBufferString(cfg)
	err = viper.ReadConfig(buf)
	if err != nil {
		t.Fatal(err)
	}

	tlist.Update("list")

	if !tlist.Exist("text") {
		t.Fatal("fail")
	}

	if tlist.Exist("bar") {
		t.Fatal("fail")
	}
}

func TestTextListpNotList(t *testing.T) {
	_, err := NewTextList("nolist")
	if err != nil && !e.Contains(err, "can't read the list from viper config") {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error is nil")
	}
}
