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

var IPprotocols = map[uint8]func([]byte) fmt.Formatter{
	IPPROTO_ICMP: func(data []byte) fmt.Formatter {
		return ICMP(data)
	},
	IPPROTO_TCP: func(data []byte) fmt.Formatter {
		return TCP(data)
	},
	IPPROTO_UDP: func(data []byte) fmt.Formatter {
		return UDP(data)
	},
}

type IP []byte

func (pdu IP) Header() (h netph.IP, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu IP) Data() (d []byte) {
	if len(pdu) >= 1 {
		n := int(pdu[0]&0xf) * 4
		if len(pdu) > n {
			d = []byte(pdu)[n:]
		}
	}
	return
}

func (pdu IP) Options() (d []byte) {
	if len(pdu) >= 1 {
		n := int(pdu[0]&0xf) * 4
		if len(pdu) > 20 {
			d = []byte(pdu)[20:n]
		}
	}
	return
}

func (pdu IP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ip ")
	h, err := pdu.Header()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	d := pdu.Data()
	fmt.Fprint(w, net.IP(h.DA[:]), " <- ", net.IP(h.SA[:]))
	if f, ok := IPprotocols[h.Protocol]; ok {
		fmt.Fprint(w, Mark, f(d))
	} else {
		fmt.Fprintf(w, ", protocol[%#x]", h.Protocol)
	}
}
