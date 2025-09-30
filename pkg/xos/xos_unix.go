// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix

package xos

import (
	"errors"
	"os"
	. "os"

	"golang.org/x/sys/unix"
)

var Termination = []Signal{Interrupt, Signal(unix.SIGTERM)}

func IsBlocked(err error) bool {
	return errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK)
}

func SetNonblockFile(f *File, val bool) error {
	return unix.SetNonblock(int(f.Fd()), val)
}

func AppendFile(name string, perm FileMode) (*File, error) {
	return os.OpenFile(name, O_RDWR|O_CREATE|O_APPEND, perm)
}

func TruncFile(name string, perm FileMode) (*File, error) {
	return os.OpenFile(name, O_RDWR|O_CREATE|O_TRUNC, perm)
}
