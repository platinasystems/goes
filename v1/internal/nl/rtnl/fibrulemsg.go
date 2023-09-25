// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"unsafe"

	"github.com/platinasystems/goes/internal/nl"
)

const SizeofFibRuleMsg = 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 4

type FibRuleMsg struct {
	Family  uint8
	Dst_len uint8
	Src_len uint8
	Tos     uint8
	Table   uint8
	_       uint8
	_       uint8
	Action  uint8
	Flags   uint32
}

func FibRuleMsgPtr(b []byte) *FibRuleMsg {
	if len(b) < nl.SizeofHdr+SizeofFibRuleMsg {
		return nil
	}
	return (*FibRuleMsg)(unsafe.Pointer(&b[nl.SizeofHdr]))
}

func (msg FibRuleMsg) Read(b []byte) (int, error) {
	*(*FibRuleMsg)(unsafe.Pointer(&b[0])) = msg
	return SizeofFibRuleMsg, nil
}

const (
	FIB_RULE_PERMANENT uint32 = 1 << iota
	FIB_RULE_INVERT
	FIB_RULE_UNRESOLVED
	FIB_RULE_IIF_DETACHED
	FIB_RULE_OIF_DETACHED
)

const FIB_RULE_DEV_DETACHED = FIB_RULE_IIF_DETACHED

const FIB_RULE_FIND_SADDR uint32 = 0x00010000

const (
	FRA_UNSPEC uint16 = iota
	FRA_DST
	FRA_SRC
	FRA_IIFNAME
	FRA_GOTO
	FRA_UNUSED2
	FRA_PRIORITY
	FRA_UNUSED3
	FRA_UNUSED4
	FRA_UNUSED5
	FRA_FWMARK
	FRA_FLOW
	FRA_TUN_ID
	FRA_SUPPRESS_IFGROUP
	FRA_SUPPRESS_PREFIXLEN
	FRA_TABLE
	FRA_FWMASK
	FRA_OIFNAME
	FRA_PAD
	FRA_L3MDEV
	FRA_UID_RANGE
	N_FRA
)

const FRA_MAX = N_FRA - 1
const FRA_IFNAME = FRA_IIFNAME

type Fra [N_FRA][]byte

func (fra *Fra) Write(b []byte) (int, error) {
	i := nl.NLMSG.Align(nl.SizeofHdr + SizeofFibRuleMsg)
	if i >= len(b) {
		nl.IndexAttrByType(fra[:], nl.Empty)
		return 0, nil
	}
	nl.IndexAttrByType(fra[:], b[i:])
	return len(b) - i, nil
}

const (
	FR_ACT_UNSPEC uint8 = iota
	FR_ACT_TO_TBL
	FR_ACT_GOTO
	FR_ACT_NOP
	FR_ACT_RES3
	FR_ACT_RES4
	FR_ACT_BLACKHOLE
	FR_ACT_UNREACHABLE
	FR_ACT_PROHIBIT
	N_FR_ACT
)

const FR_ACT_MAX = N_FR_ACT - 1

const SizeofFibRuleUidRange = 4 + 4

type FibRuleUidRange struct {
	Start uint32
	End   uint32
}

func FibRuleUidRangePtr(b []byte) *FibRuleUidRange {
	if len(b) < nl.SizeofHdr+SizeofFibRuleUidRange {
		return nil
	}
	return (*FibRuleUidRange)(unsafe.Pointer(&b[nl.SizeofHdr]))
}
