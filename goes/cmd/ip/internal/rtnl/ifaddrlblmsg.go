// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "unsafe"

const SizeofIfAddrLblMsg = 1 + 1 + 1 + 1 + 4 + 4

type IfAddrLblMsg struct {
	Family     uint8
	_          uint8
	PrefixLen  uint8
	ifal_Flags uint8
	IfIndex    uint32
	Seq        uint32
}

func IfAddrLblMsgPtr(b []byte) *IfAddrLblMsg {
	if len(b) < SizeofHdr+SizeofIfAddrLblMsg {
		return nil
	}
	return (*IfAddrLblMsg)(unsafe.Pointer(&b[SizeofHdr]))
}

func (msg IfAddrLblMsg) Read(b []byte) (int, error) {
	*(*IfAddrLblMsg)(unsafe.Pointer(&b[0])) = msg
	return SizeofIfAddrLblMsg, nil
}

const (
	IFAL_ADDRESS uint16 = 1 + iota
	IFAL_LABEL
	N_IFAL
)

const IFAL_MAX = N_IFAL - 1

type Ifal [N_IFAL][]byte

func (ifal *Ifal) Write(b []byte) (int, error) {
	i := Align(SizeofHdr + SizeofIfAddrLblMsg)
	if i >= len(b) {
		IndexAttrByType(ifal[:], Empty)
		return 0, nil
	}
	IndexAttrByType(ifal[:], b[i:])
	return len(b) - i, nil
}
