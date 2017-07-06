// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"syscall"
	"unsafe"
)

const SizeofIfInfoMsg = syscall.SizeofIfInfomsg

type IfInfoMsg syscall.IfInfomsg

func IfInfoMsgPtr(b []byte) *IfInfoMsg {
	if len(b) < SizeofHdr+SizeofIfInfoMsg {
		return nil
	}
	return (*IfInfoMsg)(unsafe.Pointer(&b[SizeofHdr]))
}

func (msg IfInfoMsg) Read(b []byte) (int, error) {
	*(*IfInfoMsg)(unsafe.Pointer(&b[0])) = msg
	return SizeofIfInfoMsg, nil
}

const (
	IFLA_UNSPEC uint16 = iota
	IFLA_ADDRESS
	IFLA_BROADCAST
	IFLA_IFNAME
	IFLA_MTU
	IFLA_LINK
	IFLA_QDISC
	IFLA_STATS
	IFLA_COST
	IFLA_PRIORITY
	IFLA_MASTER
	IFLA_WIRELESS
	IFLA_PROTINFO
	IFLA_TXQLEN
	IFLA_MAP
	IFLA_WEIGHT
	IFLA_OPERSTATE
	IFLA_LINKMODE
	IFLA_LINKINFO
	IFLA_NET_NS_PID
	IFLA_IFALIAS
	IFLA_NUM_VF
	IFLA_VFINFO_LIST
	IFLA_STATS64
	IFLA_VF_PORTS
	IFLA_PORT_SELF
	IFLA_AF_SPEC
	IFLA_GROUP
	IFLA_NET_NS_FD
	IFLA_EXT_MASK
	IFLA_PROMISCUITY
	IFLA_NUM_TX_QUEUES
	IFLA_NUM_RX_QUEUES
	IFLA_CARRIER
	IFLA_PHYS_PORT_ID
	IFLA_CARRIER_CHANGES
	IFLA_PHYS_SWITCH_ID
	IFLA_LINK_NETNSID
	IFLA_PHYS_PORT_NAME
	IFLA_PROTO_DOWN
	IFLA_GSO_MAX_SEGS
	IFLA_GSO_MAX_SIZE
	N_IFLA
)

const IFLA_MAX = N_IFLA - 1

type Ifla [N_IFLA][]byte

func (ifla *Ifla) Write(b []byte) (int, error) {
	i := Align(SizeofHdr + SizeofIfInfoMsg)
	if i >= len(b) {
		IndexAttrByType(ifla[:], Empty)
		return 0, nil
	}
	IndexAttrByType(ifla[:], b[i:])
	return len(b) - i, nil
}
