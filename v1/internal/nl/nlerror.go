// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nl

import "unsafe"

const SizeofNlmsgerr = 4 + SizeofHdr

type Nlmsgerr struct {
	// Unix errno for error.
	Errno int32
	// Header for message with error.
	Req Hdr
}

func NlmsgerrPtr(b []byte) *Nlmsgerr {
	if len(b) < SizeofHdr+SizeofNlmsgerr {
		return nil
	}
	return (*Nlmsgerr)(unsafe.Pointer(&b[SizeofHdr]))
}

func (msg Nlmsgerr) Read(b []byte) (int, error) {
	*(*Nlmsgerr)(unsafe.Pointer(&b[0])) = msg
	return SizeofNlmsgerr, nil
}
