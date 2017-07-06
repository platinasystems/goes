// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"syscall"
	"unsafe"
)

const SizeofRtGenMsg = syscall.SizeofRtGenmsg

type RtGenMsg syscall.RtGenmsg

func (msg RtGenMsg) Read(b []byte) (int, error) {
	*(*RtGenMsg)(unsafe.Pointer(&b[0])) = msg
	return SizeofRtGenMsg, nil
}

// Point to RtGenMsg w/in full message (i.e. after Hdr)
func RtGenMsgPtr(b []byte) *RtGenMsg {
	if len(b) < SizeofHdr+SizeofRtGenMsg {
		return nil
	}
	return (*RtGenMsg)(unsafe.Pointer(&b[SizeofHdr]))
}
