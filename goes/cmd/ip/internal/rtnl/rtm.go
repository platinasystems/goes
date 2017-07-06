// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

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

	RTM_NEWLINK uint16 = syscall.RTM_NEWLINK
	RTM_DELLINK uint16 = syscall.RTM_DELLINK
	RTM_GETLINK uint16 = syscall.RTM_GETLINK
	RTM_SETLINK uint16 = syscall.RTM_SETLINK

	RTM_NEWADDR uint16 = syscall.RTM_NEWADDR
	RTM_DELADDR uint16 = syscall.RTM_DELADDR
	RTM_GETADDR uint16 = syscall.RTM_GETADDR

	RTM_NEWROUTE uint16 = syscall.RTM_NEWROUTE
	RTM_DELROUTE uint16 = syscall.RTM_DELROUTE
	RTM_GETROUTE uint16 = syscall.RTM_GETROUTE

	RTM_NEWNEIGH uint16 = syscall.RTM_NEWNEIGH
	RTM_DELNEIGH uint16 = syscall.RTM_DELNEIGH
	RTM_GETNEIGH uint16 = syscall.RTM_GETNEIGH

	RTM_NEWRULE uint16 = syscall.RTM_NEWRULE
	RTM_DELRULE uint16 = syscall.RTM_DELRULE
	RTM_GETRULE uint16 = syscall.RTM_GETRULE

	RTM_NEWQDISC uint16 = syscall.RTM_NEWQDISC
	RTM_DELQDISC uint16 = syscall.RTM_DELQDISC
	RTM_GETQDISC uint16 = syscall.RTM_GETQDISC

	RTM_NEWTCLASS uint16 = syscall.RTM_NEWTCLASS
	RTM_DELTCLASS uint16 = syscall.RTM_DELTCLASS
	RTM_GETTCLASS uint16 = syscall.RTM_GETTCLASS

	RTM_NEWTFILTER uint16 = syscall.RTM_NEWTFILTER
	RTM_DELTFILTER uint16 = syscall.RTM_DELTFILTER
	RTM_GETTFILTER uint16 = syscall.RTM_GETTFILTER

	RTM_NEWACTION uint16 = syscall.RTM_NEWACTION
	RTM_DELACTION uint16 = syscall.RTM_DELACTION
	RTM_GETACTION uint16 = syscall.RTM_GETACTION

	RTM_NEWPREFIX uint16 = syscall.RTM_NEWPREFIX

	RTM_GETMULTICAST uint16 = syscall.RTM_GETMULTICAST

	RTM_GETANYCAST uint16 = syscall.RTM_GETANYCAST

	RTM_NEWNEIGHTBL uint16 = syscall.RTM_NEWNEIGHTBL
	RTM_GETNEIGHTBL uint16 = syscall.RTM_GETNEIGHTBL
	RTM_SETNEIGHTBL uint16 = syscall.RTM_SETNEIGHTBL

	RTM_NEWNDUSEROPT uint16 = syscall.RTM_NEWNDUSEROPT

	RTM_NEWADDRLABEL uint16 = syscall.RTM_NEWADDRLABEL
	RTM_DELADDRLABEL uint16 = syscall.RTM_DELADDRLABEL
	RTM_GETADDRLABEL uint16 = syscall.RTM_GETADDRLABEL

	RTM_GETDCB uint16 = syscall.RTM_GETDCB
	RTM_SETDCB uint16 = syscall.RTM_SETDCB

	RTM_NEWNETCONF uint16 = 80
	RTM_GETNETCONF uint16 = 82

	RTM_NEWMDB uint16 = 84
	RTM_DELMDB uint16 = 85
	RTM_GETMDB uint16 = 86

	RTM_NEWNSID uint16 = 88
	RTM_DELNSID uint16 = 89
	RTM_GETNSID uint16 = 90
)

var Empty = []byte{}

func NewMessage(hdr Hdr, msg io.Reader, attrs ...Attr) ([]byte, error) {
	b := make([]byte, syscall.Getpagesize())
	nmsg, err := msg.Read(b[SizeofHdr:])
	if err != nil {
		return nil, err
	}
	n := Align(SizeofHdr + nmsg)
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
