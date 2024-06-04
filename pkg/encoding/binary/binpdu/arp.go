// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binpdu

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binph"
)

type ARP []byte

func (pdu ARP) Format(w fmt.State, verb rune) {
	var h binph.ARP
	fmt.Fprint(w, "arp: ")
	if payload := h.PullFrom(pdu); len(payload) == len(pdu) {
		fmt.Fprint(w, ErrUnderrun)
	} else {
		switch h.OPER {
		case 1:
			fmt.Fprint(w, net.IP(h.SPA[:]), " request ",
				net.IP(h.TPA[:]))
		case 2:
			fmt.Fprint(w, net.IP(h.TPA[:]), " reply ",
				net.HardwareAddr(h.THA[:]))
		default:
			fmt.Fprint(w, "op[", h.OPER, "]")
		}
	}
}
