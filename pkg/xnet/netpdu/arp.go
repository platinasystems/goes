// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type ARP []byte

func (pdu ARP) Header() (h netph.ARP, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu ARP) Data() (d []byte) {
	if len(pdu) >= netph.SizeofARP {
		d = []byte(pdu)[netph.SizeofARP:]
	}
	return
}

func (pdu ARP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "arp: ")
	if h, err := pdu.Header(); err != nil {
		fmt.Fprint(w, err)
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
