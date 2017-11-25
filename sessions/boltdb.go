// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package sessions

import (
	"bytes"
	"encoding/binary"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/fcavani/e"
	"github.com/fcavani/types"
	"gopkg.in/vmihailenco/msgpack.v2"
)

// SessionBoltDB represente a session.
type SessionBoltDB struct {
	db    *bolt.DB
	codec Codec
	uuid  string
}

var bUUID = []byte("bUUID")
var bUUIDTimeStamp = []byte("bUUIDTS")
var bTimeStampUUID = []byte("bTSUUID")
var bInUse = []byte("bInUse")

func openSessionBoltDB(db *bolt.DB, codec Codec, uuid string) (Session, error) {
	var err error
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, er := tx.CreateBucketIfNotExists(bUUID)
		if er != nil {
			return e.New(er)
		}
		_, er = bucket.CreateBucketIfNotExists([]byte(uuid))
		if er != nil {
			return e.New(er)
		}
		return nil
	})
	if err != nil {
		return nil, e.Forward(err)
	}
	return &SessionBoltDB{
		db:    db,
		codec: codec,
		uuid:  uuid,
	}, nil
}

// UUID of this session.
func (s *SessionBoltDB) UUID() string {
	return s.uuid
}

// ErrCantDecData is an error.
const ErrCantDecData = "can't decode the data"

// Get restore something under the key.
func (s *SessionBoltDB) Get(key string) (interface{}, error) {
	var err error
	var val interface{}
	err = s.db.View(func(tx *bolt.Tx) error {
		var er error
		b := tx.Bucket(bUUID)
		if b == nil {
			return e.New(ErrSessionDataNotFound)
		}
		bs := b.Bucket([]byte(s.uuid))
		if bs == nil {
			return e.New(ErrSessionDataNotFound)
		}
		buf := bs.Get([]byte(key))
		if buf == nil {
			return e.New(ErrSessionDataNotFound)
		}
		val, er = s.codec.Decode(buf)
		if er != nil {
			return e.Push(er, ErrCantDecData)
		}
		return nil
	})
	if err != nil {
		return nil, e.Forward(err)
	}
	return val, nil
}

// Set store data under the name of key.
func (s *SessionBoltDB) Set(key string, data interface{}) error {
	var err error
	err = s.db.Update(func(tx *bolt.Tx) error {
		var er error
		b := tx.Bucket(bUUID)
		if b == nil {
			return e.New(ErrSessionCantStoredat)
		}
		bs, er := b.CreateBucketIfNotExists([]byte(s.uuid))
		if er != nil {
			return e.Push(er, ErrSessionCantStoredat)
		}
		buf, er := s.codec.Encode(data)
		if er != nil {
			return e.Forward(er)
		}
		er = bs.Put([]byte(key), buf)
		if er != nil {
			return e.Forward(er)
		}
		return nil
	})
	if err != nil {
		return e.Forward(err)
	}
	return nil
}

// Del deletes something under the key.
func (s *SessionBoltDB) Del(uuid string) error {
	var err error
	key := []byte(uuid)
	err = s.db.Update(func(tx *bolt.Tx) error {
		var er error
		b := tx.Bucket(bUUID)
		if b == nil {
			return e.New(ErrSessionDataNotFound)
		}
		bs := b.Bucket([]byte(s.uuid))
		if bs == nil {
			return e.New(ErrSessionDataNotFound)
		}
		buf := bs.Get(key)
		if buf == nil {
			return e.New(ErrSessionDataNotFound)
		}
		er = bs.Delete(key)
		if er != nil {
			return e.Forward(er)
		}
		if stats := bs.Stats(); stats.KeyN == 0 {
			er = b.DeleteBucket([]byte(s.uuid))
			if er != nil {
				return e.Forward(er)
			}
		}
		return nil
	})
	if err != nil {
		return e.Forward(err)
	}
	return nil
}

// Codec is the interface for encode and decode something.
type Codec interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte) (interface{}, error)
}

// MsgPackCodec implements the encoder and the decoder that is compatible with
// Codec interface.
type MsgPackCodec struct{}

// DefaultCodec if the default msgpack codec.
var DefaultCodec = MsgPackCodec{}

// Encode something.
func (m MsgPackCodec) Encode(val interface{}) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := msgpack.NewEncoder(buf)
	err := enc.EncodeString(types.Name(val))
	if err != nil {
		return nil, e.New(err)
	}
	err = enc.Encode(val)
	if err != nil {
		return nil, e.New(err)
	}
	return buf.Bytes(), nil
}

// Decode something.
func (m MsgPackCodec) Decode(b []byte) (interface{}, error) {
	buf := bytes.NewBuffer(b)
	dec := msgpack.NewDecoder(buf)
	typename, err := dec.DecodeString()
	if err != nil {
		return nil, e.New(err)
	}
	val := types.Make(types.Type(typename))
	err = dec.DecodeValue(val)
	if err != nil {
		return nil, e.New(err)
	}
	if !val.IsValid() {
		return nil, e.New("can't decode")
	}
	return val.Interface(), nil
}

// SessionsBoltDB represente a group of sessions.
type SessionsBoltDB struct {
	db    *bolt.DB
	ttl   time.Duration
	codec Codec
}

// OpenSessionsBoltDB open a session handler.
func OpenSessionsBoltDB(filename string, perm os.FileMode, opt *bolt.Options, ttl time.Duration, codec Codec) (Sessions, error) {
	var err error
	db, err := bolt.Open(filename, perm, opt)
	if err != nil {
		return nil, e.Forward(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, er := tx.CreateBucketIfNotExists(bUUIDTimeStamp)
		if er != nil {
			return e.New(er)
		}
		_, er = tx.CreateBucketIfNotExists(bTimeStampUUID)
		if er != nil {
			return e.New(er)
		}
		tx.DeleteBucket(bInUse)
		_, er = tx.CreateBucket(bInUse)
		if er != nil {
			return e.New(er)
		}
		return nil
	})
	if err != nil {
		return nil, e.Forward(err)
	}
	if codec == nil {
		codec = DefaultCodec
	}
	sb := &SessionsBoltDB{
		db:    db,
		ttl:   ttl,
		codec: codec,
	}
	sb.cleanup()
	return sb, nil
}

func (s *SessionsBoltDB) cleanup() {
	//panic("fazer isso")
}

// ErrUUIDDup is the error for duplicated uuid.
const ErrUUIDDup = "uuid duplicated"

//New creates a new session.
func (s *SessionsBoltDB) New(uuid string) (Session, error) {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bUUIDTimeStamp)
		key := []byte(uuid)
		buf := b.Get(key)
		if buf != nil {
			return e.New(ErrUUIDDup)
		}
		date := time.Now().Add(s.ttl)
		bufDate := make([]byte, binary.MaxVarintLen64)
		binary.PutVarint(bufDate, date.UnixNano())
		invertBytesOrder(bufDate)
		er := b.Put(key, bufDate)
		if er != nil {
			return e.New(er)
		}
		b = tx.Bucket(bTimeStampUUID)
		er = b.Put(bufDate, key)
		if er != nil {
			return e.New(er)
		}
		bINUSE := tx.Bucket(bInUse)
		er = bINUSE.Put(key, make([]byte, 0))
		if er != nil {
			return e.New(er)
		}
		return nil
	})
	if err != nil {
		return nil, e.Forward(err)
	}
	sess, err := openSessionBoltDB(s.db, s.codec, uuid)
	if err != nil {
		s.cancelSession(uuid)
		return nil, e.Forward(err)
	}
	return sess, nil
}

const ErrSessExpired = "session expired"

// Restore a existent session and update the ttl and in use flag.
func (s *SessionsBoltDB) Restore(uuid string) (Session, error) {
	err := s.db.Update(func(tx *bolt.Tx) error {
		bUUIDTS := tx.Bucket(bUUIDTimeStamp)
		bTSUUID := tx.Bucket(bTimeStampUUID)
		bINUSE := tx.Bucket(bInUse)

		key := []byte(uuid)
		er := bINUSE.Put(key, make([]byte, 0))
		if er != nil {
			return e.New(er)
		}
		buf := bUUIDTS.Get(key)
		if buf == nil {
			return e.New(ErrSessNotFound)
		}

		er = bTSUUID.Delete(buf)
		if er != nil {
			return e.New(er)
		}

		nano, _ := binary.Varint(invertBytes(buf))
		t := time.Unix(0, nano)
		if t.Before(time.Now()) {
			return e.New(ErrSessExpired)
		}

		date := time.Now().Add(s.ttl)
		bufDate := make([]byte, binary.MaxVarintLen64)
		binary.PutVarint(bufDate, date.UnixNano())
		invertBytesOrder(bufDate)
		er = bUUIDTS.Put(key, bufDate)
		if er != nil {
			return e.New(er)
		}
		er = bTSUUID.Put(bufDate, key)
		if er != nil {
			return e.New(er)
		}
		return nil
	})
	if err != nil {
		return nil, e.Forward(err)
	}
	sess, err := openSessionBoltDB(s.db, s.codec, uuid)
	if err != nil {
		s.cancelSession(uuid)
		return nil, e.Forward(err)
	}
	return sess, nil
}

// Delete session
func (s *SessionsBoltDB) Delete(uuid string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bUUID)
		bINUSE := tx.Bucket(bInUse)
		bUUIDTS := tx.Bucket(bUUIDTimeStamp)
		bTSUUID := tx.Bucket(bTimeStampUUID)

		key := []byte(uuid)
		buf := bINUSE.Get(key)
		if buf != nil {
			return e.New(ErrSessInUse)
		}
		b.DeleteBucket(key)

		buf = bUUIDTS.Get(key)
		if buf == nil {
			return e.New(ErrSessNotFound)
		}

		bTSUUID.Delete(buf)
		er := bUUIDTS.Delete(key)
		if er != nil {
			return e.New(er)
		}

		return nil
	})
	if err != nil {
		return e.Forward(err)
	}
	return nil
}

// Return remove session from in use pool.
func (s *SessionsBoltDB) Return(session Session) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		bINUSE := tx.Bucket(bInUse)
		bUUIDTS := tx.Bucket(bUUIDTimeStamp)
		key := []byte(session.UUID())
		buf := bUUIDTS.Get(key)
		if buf == nil {
			return e.New(ErrSessNotFound)
		}
		buf = bINUSE.Get(key)
		if buf == nil {
			return e.New(ErrSessNotFound)
		}
		er := bINUSE.Delete(key)
		if er != nil {
			return e.New(er)
		}
		return nil
	})
	if err != nil {
		return e.Forward(err)
	}
	return nil
}

func (s *SessionsBoltDB) cancelSession(uuid string) {
	s.db.Update(func(tx *bolt.Tx) error {
		bINUSE := tx.Bucket(bInUse)
		key := []byte(uuid)
		bINUSE.Delete(key)
		return nil
	})
}

// IsInUse checks is session is in use by someone.
func (s *SessionsBoltDB) IsInUse(uuid string) (bool, error) {
	err := s.db.View(func(tx *bolt.Tx) error {
		bINUSE := tx.Bucket(bInUse)
		bUUIDTS := tx.Bucket(bUUIDTimeStamp)
		key := []byte(uuid)
		buf := bUUIDTS.Get(key)
		if buf == nil {
			return e.New(ErrSessNotFound)
		}
		buf = bINUSE.Get(key)
		if buf != nil {
			return e.New(ErrSessInUse)
		}
		return nil
	})
	if e.Equal(err, ErrSessInUse) {
		return true, nil
	} else if err != nil {
		return false, e.Forward(err)
	}
	return false, nil
}

// Ttl returns the time to leave of the session.
func (s *SessionsBoltDB) Ttl(uuid string) (time.Time, error) {
	var ttl time.Time
	key := []byte(uuid)
	err := s.db.View(func(tx *bolt.Tx) error {
		bUUIDTS := tx.Bucket(bUUIDTimeStamp)
		buf := bUUIDTS.Get(key)
		if buf == nil {
			return e.New(ErrSessNotFound)
		}
		nano, _ := binary.Varint(invertBytes(buf))
		ttl = time.Unix(0, nano)
		if ttl.Before(time.Now()) {
			return e.New(ErrSessExpired)
		}
		return nil
	})
	if e.Equal(err, ErrSessExpired) {
		return ttl, e.Forward(err)
	} else if err != nil {
		return time.Time{}, e.Forward(err)
	}
	return ttl, nil
}

// Close the session handler.
func (s *SessionsBoltDB) Close() error {
	return s.db.Close()
}

func invertBytesOrder(buf []byte) {
	if buf == nil {
		return
	}
	l := len(buf)
	for i := 0; i < l/2; i++ {
		buf[i], buf[l-i-1] = buf[l-i-1], buf[i]
	}
}

func invertBytes(buf []byte) (out []byte) {
	if buf == nil {
		return
	}
	out = make([]byte, len(buf))
	copy(out, buf)
	l := len(out)
	for i := 0; i < l/2; i++ {
		out[i], out[l-i-1] = out[l-i-1], out[i]
	}
	return
}
