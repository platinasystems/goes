// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type ARP []byte

func (pdu ARP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "arp ")
	h, _, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	switch h.OPER {
	case 1:
		fmt.Fprint(w, net.IP(h.SPA[:]), " request ",
			net.IP(h.TPA[:]))
	case 2:
		fmt.Fprint(w, net.IP(h.TPA[:]), " reply ",
			net.HardwareAddr(h.THA[:]))
	default:
		fmt.Fprint(w, "op ", h.OPER)
	}
}

func (pdu ARP) Parse() (netph.ARP, []byte, error) {
	return netph.Parse[netph.ARP](pdu)
}
