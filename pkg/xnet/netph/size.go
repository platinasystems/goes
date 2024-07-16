// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import "unsafe"

const (
	SizeofARP       = int64(unsafe.Sizeof(ARP{}))
	SizeofEth       = int64(unsafe.Sizeof(Eth{}))
	SizeofHOP6      = int64(unsafe.Sizeof(HOP6{}))
	SizeofICMP      = int64(unsafe.Sizeof(ICMP{}))
	SizeofICMP6     = int64(unsafe.Sizeof(ICMP6{}))
	SizeofIEEE8021Q = int64(unsafe.Sizeof(IEEE8021Q{}))
	SizeofIP        = int64(unsafe.Sizeof(IP{}))
	SizeofIP6       = int64(unsafe.Sizeof(IP6{}))
	SizeofMPLS      = int64(unsafe.Sizeof(MPLS(0)))
	SizeofTCP       = int64(unsafe.Sizeof(TCP{}))
	SizeofTunPI     = int64(unsafe.Sizeof(TunPI{}))
	SizeofUDP       = int64(unsafe.Sizeof(UDP{}))
)
