// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package universal

import "net/netip"

type IPv6 [16]byte

func NewIPv6(b []byte) (*IPv6, []byte) { return (*IPv6)(b), b[16:] }

func (p *IPv6) Put(addr netip.Addr) {
	a16 := addr.As16()
	copy(p[:], a16[:])
}

func (p *IPv6) Value() netip.Addr { return netip.AddrFrom16(*p) }
