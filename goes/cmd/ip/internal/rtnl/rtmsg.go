// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"syscall"
	"unsafe"
)

const SizeofRtMsg = syscall.SizeofRtMsg

type RtMsg syscall.RtMsg

func RtMsgPtr(b []byte) *RtMsg {
	if len(b) < SizeofHdr+SizeofRtMsg {
		return nil
	}
	return (*RtMsg)(unsafe.Pointer(&b[SizeofHdr]))
}

func (msg RtMsg) Read(b []byte) (int, error) {
	*(*RtMsg)(unsafe.Pointer(&b[0])) = msg
	return SizeofRtMsg, nil
}

const (
	RTA_UNSPEC uint16 = iota
	RTA_DST
	RTA_SRC
	RTA_IIF
	RTA_OIF
	RTA_GATEWAY
	RTA_PRIORITY
	RTA_PREFSRC
	RTA_METRICS
	RTA_MULTIPATH
	RTA_PROTOINFO
	RTA_FLOW
	RTA_CACHEINFO
	RTA_SESSION
	RTA_MP_ALGO
	RTA_TABLE
	RTA_MARK
	RTA_MFC_STATS
	RTA_VIA
	RTA_NEWDST
	RTA_PREF
	RTA_ENCAP_TYPE
	RTA_ENCAP
	RTA_EXPIRES
	RTA_PAD
	RTA_UID
	RTA_TTL_PROPAGATE
	N_RTA
)

const RTA_MAX = N_RTA - 1

type Rta [N_RTA][]byte

func (rta *Rta) Write(b []byte) (int, error) {
	i := NLMSG.Align(SizeofHdr + SizeofRtMsg)
	if i >= len(b) {
		IndexAttrByType(rta[:], Empty)
		return 0, nil
	}
	IndexAttrByType(rta[:], b[i:])
	return len(b) - i, nil
}

const (
	RTAX_UNSPEC uint16 = iota
	RTAX_LOCK
	RTAX_MTU
	RTAX_WINDOW
	RTAX_RTT
	RTAX_RTTVAR
	RTAX_SSTHRESH
	RTAX_CWND
	RTAX_ADVMSS
	RTAX_REORDERING
	RTAX_HOPLIMIT
	RTAX_INITCWND
	RTAX_FEATURES
	RTAX_RTO_MIN
	RTAX_INITRWND
	RTAX_QUICKACK
	RTAX_CC_ALGO
	N_RTAX
)

const RTAX_MAX = N_RTAX - 1

const (
	RTAX_FEATURE_ECN uint32 = 1 << iota
	RTAX_FEATURE_SACK
	RTAX_FEATURE_TIMESTAMP
	RTAX_FEATURE_ALLFRAG
)

const RTAX_FEATURE_MASK = RTAX_FEATURE_ECN |
	RTAX_FEATURE_SACK |
	RTAX_FEATURE_TIMESTAMP |
	RTAX_FEATURE_ALLFRAG
