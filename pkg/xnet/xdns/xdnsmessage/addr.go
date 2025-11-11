// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"golang.org/x/net/dns/dnsmessage"
)

const (
	InAddrArpaSuffix = ".in-addr.arpa."
	IP6ArpaSuffix    = ".ip6.arpa."
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

// Convert [Reverse]'d address string to [netip.Addr]
func ParsePTR(s string) (addr netip.Addr) {
	var ss string
	if strings.HasSuffix(s, InAddrArpaSuffix) {
		ss = strings.TrimSuffix(s, InAddrArpaSuffix)
	} else if strings.HasSuffix(s, IP6ArpaSuffix) {
		ss = strings.TrimSuffix(s, IP6ArpaSuffix)
	} else {
		xlog.Errata.Println("invalid PTR:", s)
		return
	}
	segs := strings.Split(ss, ".")
	n := len(segs)
	if n != net.IPv4len && n != net.IPv6len {
		xlog.Errata.Println("invalid PTR segments:", ss)
		return
	}
	b := make([]byte, 0, n)
	for i := n - 1; i >= 0; i-- {
		val, err := strconv.ParseUint(segs[i], 10, 8)
		if err != nil {
			xlog.Errata.Print("invalid PTR: ", s, " (", err, ")")
			return
		}
		b = append(b, byte(val))
	}
	addr, _ = netip.AddrFromSlice(b)
	return
}

// With IPv4 address 1.2.3.4, return string "4.3.2.1.in-addr.arpa."
// Otherwise, with IPv6 address 0001:02030:4050:6070:809:1011:1213:1415,
// return "15.14.13.12.11.10.9.8.7.6.5.4.3.2.1.0.ip6.arpa."
func Reverse(addr netip.Addr) string {
	if addr.Is4() {
		a4 := addr.As4()
		return fmt.Sprintf("%d.%d.%d.%d%s",
			a4[3], a4[2], a4[1], a4[0],
			InAddrArpaSuffix)
	}
	a6 := addr.As16()
	return fmt.Sprintf(""+
		"%d.%d.%d.%d."+
		"%d.%d.%d.%d."+
		"%d.%d.%d.%d."+
		"%d.%d.%d.%d%s",
		a6[15], a6[14], a6[13], a6[12],
		a6[11], a6[10], a6[9], a6[8],
		a6[7], a6[6], a6[5], a6[4],
		a6[3], a6[2], a6[1], a6[0],
		IP6ArpaSuffix)
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
