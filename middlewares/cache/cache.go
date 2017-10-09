// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package cache

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fcavani/afero"
	"github.com/fcavani/e"
	"github.com/fcavani/log"
)

var perm os.FileMode = 0770

// ErrNoStats - error for the file thant hanven't stat information.
const ErrNoStats = "no stat for file %v"

// Storage is where the data is cached.
type Storage struct {
	root  string
	fs    afero.Fs
	close chan chan struct{}
	lck   sync.Mutex
}

// NewStorage creates a new place to store que documents in cache.
func NewStorage(root string, fs afero.Fs, ttl, periode time.Duration) *Storage {
	close := make(chan chan struct{})
	s := &Storage{
		root:  root,
		fs:    fs,
		close: close,
	}
	go func() {
		for {
			select {
			case <-time.After(periode):
				//s.lck.Lock()
				err := afero.Walk(fs, root, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return nil
					}
					if info.IsDir() {
						return nil
					}
					atime, _, _, err := statTimes(info)
					if err != nil {
						if !e.Equal(err, ErrNoStats) {
							log.Tag("cache", "cleanup").Println("statTimes:", err)
						}
						// if no stats because the afero's overlay, ignore the file...
						// BUG: make afero return the fileinfo.Sys
						return nil
					}
					if time.Since(atime) <= ttl {
						return nil
					}
					log.Tag("cache", "cleanup").DebugLevel().Println("Removing file:", path)
					err = fs.Remove(path)
					if err != nil {
						log.Tag("cache", "cleanup").Println("Remove fail:", err)
						return nil
					}
					log.Tag("cache", "cleanup").DebugLevel().Println("Removed file:", path)
					is, err := afero.IsEmpty(fs, filepath.Dir(path))
					if err != nil {
						log.Tag("cache", "cleanup").Println("IsEmpty fail:", err)
						return nil
					}
					if !is {
						return nil
					}
					dir := filepath.Dir(path)
					fs.RemoveAll(dir)
					log.Tag("cache", "cleanup").DebugLevel().Println("Removed dir:", dir)
					return nil
				})
				if err != nil {
					//s.lck.Unlock()
					continue
				}
				//s.lck.Unlock()
			case ch := <-close:
				ch <- struct{}{}
				break
			}
		}
	}()
	return s
}

// Close terminates the gorotime associetade with this storage.
func (s *Storage) Close() {
	ch := make(chan struct{})
	s.close <- ch
	<-ch
}

// Store reads from r and store under the metadata in d. h are the headers of
// the http ResponseWriter.
func (s *Storage) Store(d Document, r io.Reader, h map[string][]string) error {
	s.lck.Lock()
	defer s.lck.Unlock()
	name := d.Name()
	hash := d.RootDirHash()
	path := filepath.Join("/", s.root, hash, name)
	header := filepath.Join("/", s.root, hash, name+".header")

	//_, err := s.fs.Stat(path)
	//if err != nil

	err := s.fs.MkdirAll(filepath.Dir(path), perm)
	if err != nil {
		return e.Forward(err)
	}

	f, err := s.fs.OpenFile(path, os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return e.Forward(err)
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return e.Forward(err)
	}

	err = f.Sync()
	if err != nil {
		return e.Forward(err)
	}

	fh, err := s.fs.OpenFile(header, os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return e.Forward(err)
	}
	defer fh.Close()

	err = json.NewEncoder(fh).Encode(h)
	if err != nil {
		return e.Forward(err)
	}

	err = fh.Sync()
	if err != nil {
		return e.Forward(err)
	}

	return nil
}

// ReadTo retrives documents d and write it to w.
func (s *Storage) ReadTo(d Document, w http.ResponseWriter) error {
	s.lck.Lock()
	defer s.lck.Unlock()
	name := d.Name()
	hash := d.RootDirHash()
	path := filepath.Join("/", s.root, hash, name)
	header := filepath.Join("/", s.root, hash, name+".header")

	f, err := s.fs.Open(path)
	if err != nil {
		return e.Forward(err)
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	if err != nil {
		return e.Forward(err)
	}

	fh, err := s.fs.Open(header)
	if err != nil {
		return e.Forward(err)
	}
	defer fh.Close()

	dec := json.NewDecoder(fh)

	h := make(map[string][]string)
	err = dec.Decode(&h)
	if err != nil {
		return e.Forward(err)
	}

	for k, v := range h {
		for _, item := range v {
			w.Header().Set(k, item)
		}
	}

	return nil
}

// GetMTime retrive the mtime of document with metadata in d and
// return one document with the same metadata plus the mtime of the cache.
func (s *Storage) GetMTime(d Document) (Document, error) {
	s.lck.Lock()
	defer s.lck.Unlock()
	name := d.Name()
	hash := d.RootDirHash()
	path := filepath.Join("/", s.root, hash, name)

	fi, err := s.fs.Stat(path)
	if err != nil {
		return nil, e.Forward(err)
	}

	return d.New(fi.ModTime()), nil
}
