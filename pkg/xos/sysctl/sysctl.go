// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package sysctl

import (
	"encoding/binary"
	"os"
	_ "unsafe"

	"golang.org/x/sys/unix"
)

const Min = 4

var Sizes = []uintptr{
	4 << 10, // 4KB
	64 << 10,
	256 << 10,
	1 << 20, // 1MB
	4 << 20,
}

//go:linkname sysctl golang.org/x/sys/unix.sysctl
func sysctl(mib []int32, old *byte, oldlen *uintptr, new *byte, newlen uintptr) error

func Get(mib ...int32) ([]byte, error) {
	for try := 0; try < len(Sizes); try++ {
		n := Sizes[try]
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

func Len(data []byte) int { return int(binary.NativeEndian.Uint16(data)) }

func Type(data []byte) uint8    { return data[3] }
func Version(data []byte) uint8 { return data[2] }
