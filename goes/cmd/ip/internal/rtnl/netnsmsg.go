// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "unsafe"

const VarRunNetns = "/var/run/netns"

const SizeofNetnsMsg = 1 + 3

type NetnsMsg struct {
	Family uint8
	_      [3]uint8
}

func NetnsMsgPtr(b []byte) *NetnsMsg {
	if len(b) < SizeofHdr+SizeofNetnsMsg {
		return nil
	}
	return (*NetnsMsg)(unsafe.Pointer(&b[SizeofHdr]))
}

func (msg NetnsMsg) Read(b []byte) (int, error) {
	*(*NetnsMsg)(unsafe.Pointer(&b[0])) = msg
	return SizeofNetnsMsg, nil
}

const (
	NETNSA_NONE uint16 = iota
	NETNSA_NSID
	NETNSA_PID
	NETNSA_FD
	N_NETNSA
)

const NETNSA_MAX = N_NETNSA - 1

const NETNSA_UNASSIGNED_NSID int = -1

type Netnsa [N_NETNSA][]byte

func (netnsa *Netnsa) Write(b []byte) (int, error) {
	i := Align(SizeofHdr + SizeofNetnsMsg)
	if i >= len(b) {
		IndexAttrByType(netnsa[:], Empty)
		return 0, nil
	}
	IndexAttrByType(netnsa[:], b[i:])
	return len(b) - i, nil
}
