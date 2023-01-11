// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package universal

import "net/netip"

type IPv4 [4]byte

func NewIPv4(b []byte) (*IPv4, []byte) { return (*IPv4)(b), b[4:] }

func (p *IPv4) Put(addr netip.Addr) {
	a4 := addr.As4()
	copy(p[:], a4[:])
}

func (p *IPv4) Value() netip.Addr { return netip.AddrFrom4(*p) }
