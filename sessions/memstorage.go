// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package sessions

import (
	"sort"

	"github.com/fcavani/e"
)

type KeyFielder interface {
	KeyField() interface{}
	Less(x interface{}) bool
	Equal(x interface{}) bool
}

type Ordered []interface{}

func (o Ordered) search(x interface{}) (i int, found bool) {
	i = sort.Search(len(o), func(i int) bool {
		return !o[i].(KeyFielder).Less(x)
	})
	if i < len(o) && o[i].(KeyFielder).Equal(x) {
		// x is present at data[i]
		found = true
		return
	}
	// x is not present in data,
	// but i is the index where it would be inserted.
	return
}

const ErrItemFound = "item found"
const ErrInvData = "invalid data"

func (o *Ordered) Set(data interface{}) error {
	i, found := o.search(data.(KeyFielder).KeyField())
	if found {
		return e.New(ErrItemFound)
	}
	s := *o
	s = append(s, data)
	copy(s[i+1:], s[i:])
	s[i] = data
	*o = s
	return nil
}

const ErrNotFound = "not found"

func (o *Ordered) Del(key interface{}) error {
	i, found := o.search(key.(KeyFielder).KeyField())
	if !found {
		return e.New(ErrNotFound)
	}
	s := *o
	copy(s[i:], s[i+1:])
	s[len(s)-1] = nil // or the zero value of T
	s = s[:len(s)-1]
	*o = s
	return nil
}

type Transaction struct {
	toDelete map[interface{}]bool
}

func NewTransaction() *Transaction {
	return &Transaction{
		toDelete: make(map[interface{}]bool),
	}
}

func (t *Transaction) Del(key interface{}) error {
	if _, found := t.toDelete[key]; found {
		return e.New(ErrNotFound)
	}
	t.toDelete[key] = true
	return nil
}

func (t *Transaction) Commit(o Ordered) (Ordered, error) {
	var err error
	for k, _ := range t.toDelete {
		err = o.Del(k)
		if err != nil {
			return nil, err
		}
		delete(t.toDelete, k)
	}
	return o, nil
}

const ErrIterStop = "iter stop"

func (o *Ordered) Iter(f func(i int, data interface{}, tx *Transaction) error) error {
	var err error
	s := *o
	tx := NewTransaction()
	for i, data := range s {
		err = f(i, data, tx)
		if e.Equal(err, ErrIterStop) {
			break
		} else if err != nil {
			return e.Forward(err)
		}
	}
	s, err = tx.Commit(s)
	if err != nil {
		return e.Forward(err)
	}
	*o = s
	return nil
}
