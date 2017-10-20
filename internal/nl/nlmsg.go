// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nl

import (
	"io"
	"syscall"
)

const (
	NLMSG_NOOP    uint16 = syscall.NLMSG_NOOP
	NLMSG_ERROR   uint16 = syscall.NLMSG_ERROR
	NLMSG_DONE    uint16 = syscall.NLMSG_DONE
	NLMSG_OVERRUN uint16 = syscall.NLMSG_OVERRUN

	// User defined nlmsg_types.
	// rtmon generated timestamp
	NLMSG_TSTAMP uint16 = syscall.NLMSG_MIN_TYPE - 1
	// rtnl generated message containing NSID from out-of-band control.
	NLMSG_NSID uint16 = NLMSG_TSTAMP - 1

	NLMSG_MIN_TYPE uint16 = syscall.NLMSG_MIN_TYPE
)

var Empty = []byte{}

func NewMessage(hdr Hdr, msg io.Reader, attrs ...Attr) ([]byte, error) {
	b := make([]byte, PAGE.Size())
	nmsg, err := msg.Read(b[SizeofHdr:])
	if err != nil {
		return nil, err
	}
	n := NLMSG.Align(SizeofHdr + nmsg)
	na, err := ReadAllAttrs(b[n:], attrs...)
	if err != nil {
		return nil, err
	}
	n += na
	hdr.Len = uint32(n)
	if _, err = hdr.Read(b); err != nil {
		return nil, err
	}
	return b[:n], nil
}
