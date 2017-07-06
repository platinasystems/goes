// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"fmt"
	"syscall"
	"unsafe"
)

const SizeofHdr = syscall.NLMSG_HDRLEN

type Hdr syscall.NlMsghdr

func (hdr *Hdr) Elen() error { return fmt.Errorf("len: %d: invalid", hdr.Len) }
func (hdr *Hdr) Eseq() error { return fmt.Errorf("seq: %d: invalid", hdr.Seq) }
func (hdr *Hdr) Epid() error { return fmt.Errorf("pid: %d: invalid", hdr.Pid) }

func (hdr *Hdr) Read(b []byte) (int, error) {
	*HdrPtr(b) = *hdr
	return SizeofHdr, nil
}

func Ehdr(l int) error { return fmt.Errorf("len: %d: invalid", l) }

func HdrPtr(b []byte) *Hdr {
	if len(b) < SizeofHdr {
		return nil
	}
	return (*Hdr)(unsafe.Pointer(&b[0]))
}
