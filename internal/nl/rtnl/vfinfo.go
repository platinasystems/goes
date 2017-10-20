// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"syscall"
	"unsafe"

	"github.com/platinasystems/go/internal/nl"
)

const (
	IFLA_VF_INFO_UNSPEC uint16 = iota
	IFLA_VF_INFO
	N_IFLA_VF_INFO
)

const IFLA_VF_INFO_MAX = N_IFLA_VF_INFO - 1

const (
	IFLA_VF_UNSPEC uint16 = iota
	IFLA_VF_MAC
	IFLA_VF_VLAN
	IFLA_VF_TX_RATE
	IFLA_VF_SPOOFCHK
	IFLA_VF_LINK_STATE
	IFLA_VF_RATE
	IFLA_VF_RSS_QUERY_EN
	IFLA_VF_STATS
	IFLA_VF_TRUST
	IFLA_VF_IB_NODE_GUID
	IFLA_VF_IB_PORT_GUID
	IFLA_VF_VLAN_LIST
	N_IFLA_VF
)

const IFLA_VF_MAX = N_IFLA_VF - 1

const (
	IFLA_VF_VLAN_INFO_UNSPEC = iota
	IFLA_VF_VLAN_INFO
	N_IFLA_VF_VLAN
)

const IFLA_VF_VLAN_INFO_MAX = N_IFLA_VF_VLAN - 1

const (
	IFLA_VF_LINK_STATE_AUTO uint32 = iota
	IFLA_VF_LINK_STATE_ENABLE
	IFLA_VF_LINK_STATE_DISABLE
	N_IFLA_VF_LINK_STATE
)

const IFLA_VF_LINK_STATE_MAX = N_IFLA_VF_LINK_STATE - 1

var IflaVfLinkStateByName = map[string]uint32{
	"auto":    IFLA_VF_LINK_STATE_AUTO,
	"enable":  IFLA_VF_LINK_STATE_ENABLE,
	"disable": IFLA_VF_LINK_STATE_DISABLE,
}

var IflaVfLinkStateName = map[uint32]string{
	IFLA_VF_LINK_STATE_AUTO:    "auto",
	IFLA_VF_LINK_STATE_ENABLE:  "enable",
	IFLA_VF_LINK_STATE_DISABLE: "disable",
}

func ForEachVfInfo(b []byte, do func([]byte)) {
	nl.ForEachAttr(b, func(t uint16, val []byte) {
		if t == IFLA_VF_INFO {
			do(val)
		}
	})
}

type IflaVf [N_IFLA_VF][]byte

const SizeofIflaVfFlag = 4 + 4

type IflaVfFlag struct {
	Vf      uint32
	Setting uint32
}

func IflaVfFlagPtr(b []byte) *IflaVfFlag {
	if len(b) < SizeofIflaVfFlag {
		return nil
	}
	return (*IflaVfFlag)(unsafe.Pointer(&b[0]))
}

func (v IflaVfFlag) Read(b []byte) (int, error) {
	if len(b) < SizeofIflaVfFlag {
		return 0, syscall.EOVERFLOW
	}
	*(*IflaVfFlag)(unsafe.Pointer(&b[0])) = v
	return SizeofIflaVfFlag, nil
}

const SizeofIflaVfMac = 4 + 32

func IflaVfMacPtr(b []byte) *IflaVfMac {
	if len(b) < SizeofIflaVfMac {
		return nil
	}
	return (*IflaVfMac)(unsafe.Pointer(&b[0]))
}

type IflaVfMac struct {
	Vf  uint32
	Mac [32]byte
}

func (v *IflaVfMac) Read(b []byte) (int, error) {
	if len(b) < SizeofIflaVfMac {
		return 0, syscall.EOVERFLOW
	}
	*(*IflaVfMac)(unsafe.Pointer(&b[0])) = *v
	return SizeofIflaVfMac, nil
}

const SizeofIflaVfVlan = 4 + 4 + 4

type IflaVfVlan struct {
	Vf   uint32
	Vlan uint32 // 0 - 4095, 0 disables VLAN filter
	Qos  uint32
}

func IflaVfVlanPtr(b []byte) *IflaVfVlan {
	if len(b) < SizeofIflaVfVlan {
		return nil
	}
	return (*IflaVfVlan)(unsafe.Pointer(&b[0]))
}

func (v IflaVfVlan) Read(b []byte) (int, error) {
	if len(b) < SizeofIflaVfVlan {
		return 0, syscall.EOVERFLOW
	}
	*(*IflaVfVlan)(unsafe.Pointer(&b[0])) = v
	return SizeofIflaVfVlan, nil
}

const SizeofIflaVfTxRate = 4 + 4

type IflaVfTxRate struct {
	Vf   uint32
	Rate uint32
}

func IflaVfTxRatePtr(b []byte) *IflaVfTxRate {
	if len(b) < SizeofIflaVfTxRate {
		return nil
	}
	return (*IflaVfTxRate)(unsafe.Pointer(&b[0]))
}

func (v IflaVfTxRate) Read(b []byte) (int, error) {
	if len(b) < SizeofIflaVfTxRate {
		return 0, syscall.EOVERFLOW
	}
	*(*IflaVfTxRate)(unsafe.Pointer(&b[0])) = v
	return SizeofIflaVfTxRate, nil
}

const SizeofIflaVfRate = 4 + 4 + 4

type IflaVfRate struct {
	Vf        uint32
	MinTxRate uint32
	MaxTxRate uint32
}

func IflaVfRatePtr(b []byte) *IflaVfRate {
	if len(b) < SizeofIflaVfRate {
		return nil
	}
	return (*IflaVfRate)(unsafe.Pointer(&b[0]))
}

func (v IflaVfRate) Read(b []byte) (int, error) {
	if len(b) < SizeofIflaVfRate {
		return 0, syscall.EOVERFLOW
	}
	*(*IflaVfRate)(unsafe.Pointer(&b[0])) = v
	return SizeofIflaVfRate, nil
}

const SizeofIflaVfLinkState = 4 + 4

type IflaVfLinkState struct {
	Vf        uint32
	LinkState uint32
}

func IflaVfLinkStatePtr(b []byte) *IflaVfLinkState {
	if len(b) < SizeofIflaVfLinkState {
		return nil
	}
	return (*IflaVfLinkState)(unsafe.Pointer(&b[0]))
}

func (v IflaVfLinkState) Read(b []byte) (int, error) {
	if len(b) < SizeofIflaVfLinkState {
		return 0, syscall.EOVERFLOW
	}
	*(*IflaVfLinkState)(unsafe.Pointer(&b[0])) = v
	return SizeofIflaVfLinkState, nil
}

const SizeofIflaVfGuid = 4 + 8

type IflaVfGuid struct {
	Vf   uint32
	Guid uint64
}

func IflaVfIbGuidPtr(b []byte) *IflaVfGuid {
	if len(b) < SizeofIflaVfGuid {
		return nil
	}
	return (*IflaVfGuid)(unsafe.Pointer(&b[0]))
}

func (v IflaVfGuid) Read(b []byte) (int, error) {
	if len(b) < SizeofIflaVfGuid {
		return 0, syscall.EOVERFLOW
	}
	*(*IflaVfGuid)(unsafe.Pointer(&b[0])) = v
	return SizeofIflaVfGuid, nil
}
