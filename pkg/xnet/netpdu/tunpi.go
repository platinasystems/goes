// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type TunPI []byte

func (pdu TunPI) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "tun")
	h, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, " ", err)
		return
	}
	switch h.Proto {
	case netph.TUN_P_IP:
		fmt.Fprint(w, Mark, IP(d))
	case netph.TUN_P_IP6:
		fmt.Fprint(w, Mark, IP6(d))
	default:
		fmt.Fprintf(w, " proto %#x", h.Proto)
	}
}

func (pdu TunPI) Parse() (netph.TunPI, []byte, error) {
	return netph.Parse[netph.TunPI](pdu)
}

func (pdu TunPI) Proto(is6 bool) {
	ethp := uint16(netph.ETH_P_IP)
	if is6 {
		ethp = netph.ETH_P_IPV6
	}
	xnet.Encode(pdu, netph.TunPI{
		Proto: ethp,
	})
}

func (pdu TunPI) ToWhom() (addr netip.Addr, err error) {
	h, d, err := pdu.Parse()
	if err != nil {
		return
	}
	switch h.Proto {
	case netph.TUN_P_IP:
		var ip netph.IP
		if ip, _, err = IP(d).Parse(); err == nil {
			addr.UnmarshalBinary(ip.DA[:])
		}
	case netph.TUN_P_IP6:
		var ip6 netph.IP6
		if ip6, _, err = IP6(d).Parse(); err == nil {
			addr.UnmarshalBinary(ip6.DA[:])
		}
	}
	return
}
