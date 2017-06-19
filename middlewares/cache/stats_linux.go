// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

// +build linux

package cache

import (
	"os"
	"syscall"
	"time"

	"github.com/fcavani/e"
)

func statTimes(fi os.FileInfo) (atime, mtime, ctime time.Time, err error) {
	mtime = fi.ModTime()
	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		err = e.New(ErrNoStats, fi.Name())
		return
	}

	atime = time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
	ctime = time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
	return
}
