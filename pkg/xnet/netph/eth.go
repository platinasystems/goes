// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import "unsafe"

const ETHER_ADDR_LEN = 6

// https://en.wikipedia.org/wiki/Ethernet_frame
type Eth struct {
	DMAC [ETHER_ADDR_LEN]byte
	SMAC [ETHER_ADDR_LEN]byte
	Type uint16
}

const EthSize = int(unsafe.Sizeof(Eth{}))

const ETHMTU = 1500

const (
	ETH_P_8021Q   = 0x8100
	ETH_P_8021AD  = 0x88a8
	ETH_P_ARP     = 0x806
	ETH_P_IP      = 0x800
	ETH_P_IPV6    = 0x86dd
	ETH_P_MPLS_UC = 0x8847
	ETH_P_MPLS_MC = 0x8848
)

func (p *Eth) IsUnicast() bool   { return (p.DMAC[0] & 1) == 0 }
func (p *Eth) ShouldLearn() bool { return (p.SMAC[0] & 1) == 0 }

func EA64(ea [ETHER_ADDR_LEN]byte) uint64 {
	ea64 := uint64(ea[0]) << 40
	ea64 |= uint64(ea[1]) << 32
	ea64 |= uint64(ea[2]) << 24
	ea64 |= uint64(ea[3]) << 16
	ea64 |= uint64(ea[4]) << 8
	ea64 |= uint64(ea[5])
	return ea64
}
