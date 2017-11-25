// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package sessions

import "time"

type Sessions interface {
	New(uuid string) (Session, error)
	Restore(uuid string) (Session, error)
	Delete(uuid string) error
	Return(s Session) error
	IsInUse(uuid string) (bool, error)
	Ttl(uuid string) (time.Time, error)
	Close() error
}

type Session interface {
	UUID() string
	Get(key string) (interface{}, error)
	Set(key string, data interface{}) error
	Del(key string) error
}
