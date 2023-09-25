// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nl

import "syscall"

// Pop next message from front of buffer and return rem[ainder].
func Pop(b []byte) (msg, rem []byte, err error) {
	msg = Empty
	rem = Empty
	if len(b) < SizeofHdr {
		return
	}
	h := HdrPtr(b)
	n := int(h.Len)
	if n > len(b) {
		err = syscall.EOVERFLOW
		return
	}
	msg = b[:n]
	n = NLMSG.Align(n)
	if n > len(b) {
		rem = Empty
	} else {
		rem = b[n:]
	}
	return
}
