// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/internal/nl"
	"github.com/platinasystems/goes/internal/sizeof"
)

const SizeofIfAddrMsg = syscall.SizeofIfAddrmsg

type IfAddrMsg syscall.IfAddrmsg

func IfAddrMsgPtr(b []byte) *IfAddrMsg {
	if len(b) < nl.SizeofHdr+SizeofIfAddrMsg {
		return nil
	}
	return (*IfAddrMsg)(unsafe.Pointer(&b[nl.SizeofHdr]))
}

func (msg IfAddrMsg) Read(b []byte) (int, error) {
	*(*IfAddrMsg)(unsafe.Pointer(&b[0])) = msg
	return SizeofIfAddrMsg, nil
}

const (
	IFA_UNSPEC uint16 = iota
	IFA_ADDRESS
	IFA_LOCAL
	IFA_LABEL
	IFA_BROADCAST
	IFA_ANYCAST
	IFA_CACHEINFO
	IFA_MULTICAST
	IFA_FLAGS
	N_IFA
)

const IFA_MAX = N_IFA - 1

type Ifa [N_IFA][]byte

func (ifa *Ifa) Write(b []byte) (int, error) {
	i := nl.NLMSG.Align(nl.SizeofHdr + SizeofIfAddrMsg)
	if i >= len(b) {
		nl.IndexAttrByType(ifa[:], nl.Empty)
		return 0, nil
	}
	nl.IndexAttrByType(ifa[:], b[i:])
	return len(b) - i, nil
}

const SizeofIfaCacheInfo = 4 * sizeof.Long

type IfaCacheInfo struct {
	Prefered uint32
	Valid    uint32
	// timestamps are hundredths of seconds
	CreatedTimestamp uint32
	UpdatedTimestamp uint32
}

func IfaCacheInfoPtr(b []byte) *IfaCacheInfo {
	if len(b) < SizeofIfaCacheInfo {
		return nil
	}
	return (*IfaCacheInfo)(unsafe.Pointer(&b[0]))
}
