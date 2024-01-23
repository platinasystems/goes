// Derrived from golang.org/x/net/route
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package sysctl

import (
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

const MsgMin = 4

func MsgLen(data []byte) int {
	return int(*((*uint16)(unsafe.Pointer(&data[0]))))
}

func MsgType(data []byte) uint8 { return data[3] }

func MsgOK(data []byte, types ...uint8) bool {
	if data[2] == unix.RTM_VERSION {
		for _, t := range types {
			if MsgType(data) == t {
				return true
			}
		}
	}
	return false
}

var trysize = []uintptr{
	4 << 10, // 4KB
	64 << 10,
	256 << 10,
	1 << 20, // 1MB
	4 << 20,
}

func Get(mib ...int32) ([]byte, error) {
	for try := 0; try < len(trysize); try++ {
		n := trysize[try]
		b := make([]byte, n)
		if err := sysctl(mib, &b[0], &n, nil, 0); err == nil {
			return b[:n], nil
		} else if err != unix.ENOMEM {
			return nil, os.NewSyscallError("sysctl", err)
		}
		b = b[:0]
	}
	return nil, unix.ENOMEM
}
