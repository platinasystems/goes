// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
)

// ISS = Internel Shared SRAM ?
// This is the shared SRAM used for l3_defip_alpm table and others.
// ISS is physically 4 banks x 8k buckets per bank x 420 bits per bucket.
const (
	n_iss_buckets_per_bank = 8 << 10
	n_iss_banks            = 4
	n_iss_bits_per_bucket  = 420
)

// 4 bit chip internal priority
type internal_priority uint8

func (x *internal_priority) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+3, i, isSet)
}

// 2 bit chip internal congestion state
type internal_congestion_state uint8

func (x *internal_congestion_state) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

// 3 bit vlan priority
type dot1q_priority uint8

func (x *dot1q_priority) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+2, i, isSet)
}

type spanning_tree_state uint8

const (
	spanning_tree_state_disable spanning_tree_state = iota
	spanning_tree_state_blocking
	spanning_tree_state_learning
	spanning_tree_state_forwarding
)

func (x *spanning_tree_state) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

// Packet color (based on classification and metering).
type packet_color int8

const (
	packet_color_green packet_color = iota
	packet_color_yellow
	packet_color_red
	n_packet_color
)

type ip_dscp uint8

func (x *ip_dscp) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+5, i, isSet)
}

// ip header ECN bits
type ip_ecn_bits uint8

func (x *ip_ecn_bits) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}
