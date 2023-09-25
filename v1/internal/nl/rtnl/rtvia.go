// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"syscall"
	"unsafe"
)

type RtVia struct {
	Family  uint16
	Address []byte
}

func (v RtVia) Read(b []byte) (int, error) {
	if len(b) < 2+len(v.Address) {
		return 0, syscall.EOVERFLOW
	}
	*(*uint16)(unsafe.Pointer(&b[0])) = uint16(v.Family)
	return 2 + copy(b[2:], v.Address), nil
}
