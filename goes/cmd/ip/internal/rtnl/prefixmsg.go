// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "unsafe"

const SizeofPrefixMsg = 1 + 1 + 1 + 1 + SizeofInt + 1 + 1 + 1 + 1

type PrefixMsg struct {
	Family  uint8
	_       uint8
	_       uint8
	_       uint8
	IfIndex int
	Type    uint8
	Len     uint8
	Flags   uint8
	_       uint8
}

func PrefixMsgPtr(b []byte) *PrefixMsg {
	if len(b) < SizeofHdr+SizeofPrefixMsg {
		return nil
	}
	return (*PrefixMsg)(unsafe.Pointer(&b[SizeofHdr]))
}

func (msg PrefixMsg) Read(b []byte) (int, error) {
	*(*PrefixMsg)(unsafe.Pointer(&b[0])) = msg
	return SizeofPrefixMsg, nil
}

const (
	PREFIX_UNSPEC uint16 = iota
	PREFIX_ADDRESS
	PREFIX_CACHEINFO
	N_PREFIX
)

const PREFIX_MAX = N_PREFIX - 1

type Prefixa [N_PREFIX][]byte

func (prefixa *Prefixa) Write(b []byte) (int, error) {
	i := Align(SizeofHdr + SizeofPrefixMsg)
	if i >= len(b) {
		IndexAttrByType(prefixa[:], Empty)
		return 0, nil
	}
	IndexAttrByType(prefixa[:], b[i:])
	return len(b) - i, nil
}

const SizeofPrefixCacheInfo = 4 + 4

type PrefixCacheInfo struct {
	PreferredTime uint32
	ValidTime     uint32
}

func PrefixCacheInfoPtr(b []byte) *PrefixCacheInfo {
	if len(b) < SizeofHdr+SizeofPrefixCacheInfo {
		return nil
	}
	return (*PrefixCacheInfo)(unsafe.Pointer(&b[SizeofHdr]))
}

// Prefix Flags
const (
	IF_PREFIX_ONLINK   uint8 = 0x01
	IF_PREFIX_AUTOCONF uint8 = 0x02
)
