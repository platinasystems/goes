// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import "unsafe"

var ProtoMark = "; " // or \n\t

type Headers interface {
	ARP | ETH | Hop6 | ICMP | ICMP6 | IEEE8021Q | IPv4 | IPv6 |
		MPLS | BOS | TCP | UDP | TapPI | TunPI | TunTapPI
}

// This returns an unsafe reference to header structures.
func Header[T Headers](data []byte) *T {
	return (*T)(unsafe.Pointer(&data[0]))
}

// This returns an unsafe pointer to whatever is after the header structure.
func Data[T Headers](h *T) unsafe.Pointer {
	return unsafe.Add(unsafe.Pointer(h), Sizeof(h))
}

func Sizeof[T Headers](h *T) int { return int(unsafe.Sizeof(*h)) }
