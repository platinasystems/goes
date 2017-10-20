// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"unsafe"

	"github.com/platinasystems/go/internal/nl"
)

const SizeofNetconfMsg = 1

type NetconfMsg struct {
	Family uint8
}

func (msg NetconfMsg) Read(b []byte) (int, error) {
	*(*NetconfMsg)(unsafe.Pointer(&b[0])) = msg
	return SizeofNetconfMsg, nil
}

func NetconfMsgPtr(b []byte) *NetconfMsg {
	if len(b) < nl.SizeofHdr+SizeofNetconfMsg {
		return nil
	}
	return (*NetconfMsg)(unsafe.Pointer(&b[nl.SizeofHdr]))
}

const (
	NETCONFA_UNSPEC uint16 = iota
	NETCONFA_IFINDEX
	NETCONFA_FORWARDING
	NETCONFA_RP_FILTER
	NETCONFA_MC_FORWARDING
	NETCONFA_PROXY_NEIGH
	NETCONFA_IGNORE_ROUTES_WITH_LINKDOWN
	NETCONFA_INPUT
	N_NETCONFA
)

const NETCONFA_MAX = N_NETCONFA - 1

const (
	NETCONFA_IFINDEX_ALL     int32 = -1
	NETCONFA_IFINDEX_DEFAULT int32 = -2
)

type Netconfa [N_NETCONFA][]byte

func (netconfa *Netconfa) Write(b []byte) (int, error) {
	i := nl.NLMSG.Align(nl.SizeofHdr + SizeofNetconfMsg)
	if i >= len(b) {
		nl.IndexAttrByType(netconfa[:], nl.Empty)
		return 0, nil
	}
	nl.IndexAttrByType(netconfa[:], b[i:])
	return len(b) - i, nil
}
