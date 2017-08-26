// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package sessions

import (
	"testing"

	"github.com/fcavani/e"
)

type testtype string

func (t testtype) KeyField() interface{} {
	return t
}

func (t testtype) Less(x interface{}) bool {
	s := string(x.(testtype))
	return string(t) < s
}

func (t testtype) Equal(x interface{}) bool {
	s := string(x.(testtype))
	return string(t) == s
}

var o Ordered

func init() {
	o = make(Ordered, 0)
}

func TestSet(t *testing.T) {
	err := o.Set(testtype("1"))
	if err != nil {
		t.Fatal(err)
	}
	err = o.Set(testtype("1"))
	if err != nil && !e.Equal(err, ErrItemFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error nil")
	}
	err = o.Set(testtype("2"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestDel(t *testing.T) {
	err := o.Del(testtype("1"))
	if err != nil {
		t.Fatal(err)
	}
	err = o.Del(testtype("1"))
	if err != nil && !e.Equal(err, ErrNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error nil")
	}
}

func TestIter(t *testing.T) {
	err := o.Set(testtype("1"))
	if err != nil {
		t.Fatal(err)
	}
	err = o.Set(testtype("3"))
	if err != nil {
		t.Fatal(err)
	}

	err = o.Iter(func(i int, data interface{}, tx *Transaction) error {
		t.Log(i, data)
		er := tx.Del(data)
		if er != nil {
			t.Fatal(er)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = o.Del(testtype("1"))
	if err != nil && !e.Equal(err, ErrNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error nil")
	}
	err = o.Del(testtype("2"))
	if err != nil && !e.Equal(err, ErrNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error nil")
	}
	err = o.Del(testtype("3"))
	if err != nil && !e.Equal(err, ErrNotFound) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("error nil")
	}

	err = o.Set(testtype("1"))
	if err != nil {
		t.Fatal(err)
	}
	err = o.Set(testtype("2"))
	if err != nil {
		t.Fatal(err)
	}
	err = o.Set(testtype("3"))
	if err != nil {
		t.Fatal(err)
	}

	itens := make(Ordered, 0)

	err = o.Iter(func(i int, data interface{}, tx *Transaction) error {
		t.Log(i, data)
		itens = append(itens, data)
		if i == 1 {
			return e.New(ErrIterStop)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(itens) != 2 {
		t.Fatal("iter fail", len(itens))
	}

}
