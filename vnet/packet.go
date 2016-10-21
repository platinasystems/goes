// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

// Generic packet type
type PacketType int

const (
	IP4 PacketType = 1 + iota
	IP6
	MPLS_UNICAST
	MPLS_MULTICAST
	ARP
)

// Interface defines SupportsArp method to enable glean adjacencies.
type Arper interface {
	SupportsArp()
}
