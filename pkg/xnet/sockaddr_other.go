// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build android || illumos || linux || solaris || windows

package xnet

import (
	"encoding/binary"

	"golang.org/x/sys/unix"
)

func (b SAInBuf) Family() int {
	return int(binary.NativeEndian.Uint16(b[:2]))
}

func (b SAInBuf) SetINET() {
	binary.NativeEndian.PutUint16(b[:2], unix.AF_INET)
}

func (b SAInBuf) SetINET6() {
	binary.NativeEndian.PutUint16(b[:2], unix.AF_INET6)
}
