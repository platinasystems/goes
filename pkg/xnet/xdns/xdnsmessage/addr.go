// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"fmt"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

type TypeAResource struct{ netip.Addr }
type TypeAAAAResource struct{ netip.Addr }

type AddrResources interface {
	TypeAResource | TypeAAAAResource
}

func ParseAddr[T AddrResources](tokens []string) (T, error) {
	if len(tokens) == 0 {
		return T{netip.Addr{}}, xerrors.Incomplete("A|AAAA")
	}
	addr, err := netip.ParseAddr(tokens[0])
	return T{addr}, err
}

// With IPv4 address 1.2.3.4, return string "3.2.1.0.in-addr.arpa."
// Otherwise, with IPv6 address 0001:02030:4050:6070:809:1011:1213:1415,
// return "15.14.13.12.11.10.9.8.7.6.5.4.3.2.1.0.ip6.arpa."
func Reverse(addr netip.Addr) string {
	if addr.Is4() {
		a4 := addr.As4()
		return fmt.Sprintf("%d.%d.%d.%d.in-addr.arpa.",
			a4[3], a4[2], a4[1], a4[0])
	}
	a6 := addr.As16()
	return fmt.Sprintf(""+
		"%d.%d.%d.%d."+
		"%d.%d.%d.%d."+
		"%d.%d.%d.%d."+
		"%d.%d.%d.%d."+
		"ip6.arpa.",
		a6[15], a6[14], a6[13], a6[12],
		a6[11], a6[10], a6[9], a6[8],
		a6[7], a6[6], a6[5], a6[4],
		a6[3], a6[2], a6[1], a6[0])
}

func (v TypeAResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	var a dnsmessage.AResource
	copy(a.A[:], v.AsSlice())
	return mb.AResource(h, a)
}

func (v TypeAAAAResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	var aaaa dnsmessage.AAAAResource
	copy(aaaa.AAAA[:], v.AsSlice())
	return mb.AAAAResource(h, aaaa)
}
