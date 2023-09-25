// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package omni

import "net/netip"

type IPv6 [16]byte

func (ip *IPv6) Put(addr netip.Addr) {
	a16 := addr.As16()
	copy(ip[:], a16[:])
}

func (ip *IPv6) Value() netip.Addr { return netip.AddrFrom16(*ip) }
