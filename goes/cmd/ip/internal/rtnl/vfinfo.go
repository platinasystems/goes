// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "unsafe"

const (
	IFLA_VF_INFO_UNSPEC = iota
	IFLA_VF_INFO
	N_IFLA_VF_INFO
)

const IFLA_VF_INFO_MAX = N_IFLA_VF_INFO - 1

const (
	IFLA_VF_UNSPEC = iota
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
	IFLA_VF_LINK_STATE_AUTO = iota
	IFLA_VF_LINK_STATE_ENABLE
	IFLA_VF_LINK_STATE_DISABLE
	N_IFLA_VF_LINK_STATE
)

const IFLA_VF_LINK_STATE_MAX = N_IFLA_VF_LINK_STATE - 1

func ForEachVfInfo(b []byte, do func([]byte)) {
	ForEachAttr(b, func(t uint16, val []byte) {
		if t == IFLA_VF_INFO {
			do(val)
		}
	})
}

type IflaVf [N_IFLA_VF][]byte

const SizeofIflaVfMac = 4 + 32

func IflaVfMacPtr(vf *IflaVf) *IflaVfMac {
	if len(vf[IFLA_VF_MAC]) < SizeofIflaVfMac {
		return nil
	}
	return (*IflaVfMac)(unsafe.Pointer(&vf[IFLA_VF_MAC][0]))
}

type IflaVfMac struct {
	Vf  uint32
	Mac [32]byte
}

const SizeofIflaVfVlan = 4 + 4 + 4

type IflaVfVlan struct {
	Vf   uint32
	Vlan uint32 // 0 - 4095, 0 disables VLAN filter
	Qos  uint32
}

func IflaVfVlanPtr(vf *IflaVf) *IflaVfVlan {
	if len(vf[IFLA_VF_VLAN]) < SizeofIflaVfVlan {
		return nil
	}
	return (*IflaVfVlan)(unsafe.Pointer(&vf[IFLA_VF_VLAN][0]))
}

const SizeofIflaVfTxRate = 4 + 4

type IflaVfTxRate struct {
	Vf   uint32
	Rate uint32
}

func IflaVfTxRatePtr(vf *IflaVf) *IflaVfTxRate {
	if len(vf[IFLA_VF_TX_RATE]) < SizeofIflaVfTxRate {
		return nil
	}
	return (*IflaVfTxRate)(unsafe.Pointer(&vf[IFLA_VF_TX_RATE][0]))
}

const SizeofIflaVfRate = 4 + 4 + 4

type IflaVfRate struct {
	Vf        uint32
	MinTxRate uint32
	MaxTxRate uint32
}

func IflaVfRatePtr(vf *IflaVf) *IflaVfRate {
	if len(vf[IFLA_VF_RATE]) < SizeofIflaVfRate {
		return nil
	}
	return (*IflaVfRate)(unsafe.Pointer(&vf[IFLA_VF_RATE][0]))
}

const SizeofIflaVfSpoofchk = 4 + 4

type IflaVfSpoofchk struct {
	Vf      uint32
	Setting uint32
}

func IflaVfSpoofchkPtr(vf *IflaVf) *IflaVfSpoofchk {
	if len(vf[IFLA_VF_SPOOFCHK]) < SizeofIflaVfSpoofchk {
		return nil
	}
	return (*IflaVfSpoofchk)(unsafe.Pointer(&vf[IFLA_VF_SPOOFCHK][0]))
}

const SizeofIflaVfGuid = 4 + 8

type IflaVfGuid struct {
	Vf   uint32
	Guid uint64
}

func IflaVfIbNodeGuidPtr(vf *IflaVf) *IflaVfGuid {
	if len(vf[IFLA_VF_IB_NODE_GUID]) < SizeofIflaVfGuid {
		return nil
	}
	return (*IflaVfGuid)(unsafe.Pointer(&vf[IFLA_VF_IB_NODE_GUID][0]))
}

func IflaVfIbPortGuidPtr(vf *IflaVf) *IflaVfGuid {
	if len(vf[IFLA_VF_IB_PORT_GUID]) < SizeofIflaVfGuid {
		return nil
	}
	return (*IflaVfGuid)(unsafe.Pointer(&vf[IFLA_VF_IB_PORT_GUID][0]))
}

const SizeofIflaVfLinkState = 4 + 4

type IflaVfLinkState struct {
	Vf        uint32
	LinkState uint32
}

func IflaVfLinkStatePtr(vf *IflaVf) *IflaVfLinkState {
	if len(vf[IFLA_VF_LINK_STATE]) < SizeofIflaVfLinkState {
		return nil
	}
	return (*IflaVfLinkState)(unsafe.Pointer(&vf[IFLA_VF_LINK_STATE][0]))
}

const SizeofIflaVfTrust = 4 + 4

type IflaVfTrust struct {
	Vf      uint32
	Setting uint32
}

func IflaVfTrustPtr(vf *IflaVf) *IflaVfTrust {
	if len(vf[IFLA_VF_TRUST]) < SizeofIflaVfTrust {
		return nil
	}
	return (*IflaVfTrust)(unsafe.Pointer(&vf[IFLA_VF_TRUST][0]))
}
