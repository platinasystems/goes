// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"unsafe"

	"github.com/platinasystems/goes/internal/nl"
	"github.com/platinasystems/goes/internal/sizeof"
)

const SizeofNdMsg = (2 * sizeof.Byte) + sizeof.Short + sizeof.Long +
	sizeof.Short + (2 * sizeof.Byte)

type NdMsg struct {
	Family uint8
	_      uint8
	_      uint16
	Index  int32
	State  uint16
	Flags  uint8
	Type   uint8
}

func NdMsgPtr(b []byte) *NdMsg {
	if len(b) < nl.SizeofHdr+SizeofNdMsg {
		return nil
	}
	return (*NdMsg)(unsafe.Pointer(&b[nl.SizeofHdr]))
}

func (msg NdMsg) Read(b []byte) (int, error) {
	*(*NdMsg)(unsafe.Pointer(&b[0])) = msg
	return SizeofNdMsg, nil
}

const (
	NDA_UNSPEC uint16 = iota
	NDA_DST
	NDA_LLADDR
	NDA_CACHEINFO
	NDA_PROBES
	NDA_VLAN
	NDA_PORT
	NDA_VNI
	NDA_IFINDEX
	NDA_MASTER
	NDA_LINK_NETNSID
	N_NDA
)

const NDA_MAX = N_NDA - 1

type Nda [N_NDA][]byte

func (nda *Nda) Write(b []byte) (int, error) {
	i := nl.NLMSG.Align(nl.SizeofHdr + SizeofNdMsg)
	if i >= len(b) {
		nl.IndexAttrByType(nda[:], nl.Empty)
		return 0, nil
	}
	nl.IndexAttrByType(nda[:], b[i:])
	return len(b) - i, nil
}

const SizeofNdaCacheInfo = 4 * sizeof.Long

type NdaCacheInfo struct {
	Confirmed uint32
	Used      uint32
	Updated   uint32
	RefCnt    uint32
}

func NdaCacheInfoPtr(b []byte) *NdaCacheInfo {
	if len(b) < SizeofNdaCacheInfo {
		return nil
	}
	return (*NdaCacheInfo)(unsafe.Pointer(&b[0]))
}
