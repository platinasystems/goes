// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package omni

import "net/netip"

type IPv4 [4]byte

func (ip *IPv4) Put(addr netip.Addr) {
	a4 := addr.As4()
	copy(ip[:], a4[:])
}

func (ip *IPv4) Value() netip.Addr { return netip.AddrFrom4(*ip) }
