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
	"regexp"
	"sort"
	"sync"

	"github.com/fcavani/e"
	log "github.com/fcavani/slog"
	"github.com/spf13/viper"
)

var MinListSize = 50

type List interface {
	Exist(str string) bool
	Update(key string)
}

type Regexp struct {
	r   []*regexp.Regexp
	lck sync.Mutex
}

func NewRegexpList(key string) (*Regexp, error) {
	slice := viper.GetStringSlice(key)
	if slice == nil || len(slice) == 0 {
		return nil, e.New("can't read the list from viper config")
	}

	rs := make([]*regexp.Regexp, 0, MinListSize)

	for _, data := range slice {
		r, err := regexp.Compile(data)
		if err != nil {
			return nil, e.Forward(err)
		}
		rs = append(rs, r)
	}

	return &Regexp{
		r: rs,
	}, nil
}

func (r *Regexp) Update(key string) {
	r.lck.Lock()
	defer r.lck.Unlock()

	slice := viper.GetStringSlice(key)
	if slice == nil || len(slice) == 0 {
		log.Error("can't read the list from viper config")
	}

	r.r = make([]*regexp.Regexp, 0, len(r.r))

	for _, data := range slice {
		re, err := regexp.Compile(data)
		if err != nil {
			log.Errorf("Fail to compile regexp expression: %v", err)
			return
		}
		r.r = append(r.r, re)
	}
}

func (r *Regexp) Exist(str string) bool {
	r.lck.Lock()
	defer r.lck.Unlock()
	for _, re := range r.r {
		if re.MatchString(str) {
			return true
		}
	}
	return false
}

type TextList struct {
	list []string
	lck  sync.Mutex
}

func NewTextList(key string) (*TextList, error) {
	slice := viper.GetStringSlice(key)
	if slice == nil || len(slice) == 0 {
		return nil, e.New("can't read the list from viper config")
	}
	sort.Strings(slice)
	return &TextList{
		list: slice,
	}, nil
}

func (t *TextList) Update(key string) {
	t.lck.Lock()
	defer t.lck.Unlock()
	slice := viper.GetStringSlice(key)
	if slice == nil || len(slice) == 0 {
		log.Error("can't read the list from viper config")
		return
	}
	sort.Strings(slice)
	t.list = slice
}

func (t *TextList) Exist(str string) bool {
	t.lck.Lock()
	defer t.lck.Unlock()
	i := sort.SearchStrings(t.list, str)
	if i >= 0 && i < len(t.list) {
		if t.list[i] == str {
			return true
		}
	}
	return false
}
