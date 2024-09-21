// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package netph provides internet protocol-header types.
package netph

import "encoding/binary"

type Header interface {
	ARP |
		Eth |
		HOP6 |
		ICMP |
		ICMP6 |
		ICMP6EchoRequest |
		ICMP6EchoReply |
		ICMP6RouterSolicitation |
		ICMP6RouterAdvertisement |
		ICMP6NeighborSolicitation |
		ICMP6NeighborAdvertisement |
		ICMP6RedirectMessage |
		IEEE8021 |
		IP |
		IP6 |
		MPLS |
		TCP |
		TunPI |
		UDP
}

func Parse[H Header](b []byte) (h H, d []byte, err error) {
	n, err := binary.Decode(b, binary.BigEndian, &h)
	if err == nil {
		d = b[n:]
	}
	return
}
