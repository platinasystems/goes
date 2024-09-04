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

var IP6NextHeaders = map[uint8]func([]byte) fmt.Formatter{
	0: func(data []byte) fmt.Formatter {
		return HOP6(data)
	},
	IPPROTO_ICMPV6: func(data []byte) fmt.Formatter {
		return ICMP6(data)
	},
	IPPROTO_TCP: func(data []byte) fmt.Formatter {
		return TCP(data)
	},
	IPPROTO_UDP: func(data []byte) fmt.Formatter {
		return UDP(data)
	},
}

type IP6 []byte

func (pdu IP6) Header() (h netph.IP6, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu IP6) Data() (d []byte) {
	if len(pdu) >= netph.SizeofIP6 {
		d = []byte(pdu)[netph.SizeofIP6:]
	}
	return
}

func (pdu IP6) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ip6 ")
	h, err := pdu.Header()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	d := pdu.Data()
	fmt.Fprint(w, net.IP(h.DA[:]), " <- ", net.IP(h.SA[:]))
	if f, ok := IP6NextHeaders[h.NextHeader]; ok {
		fmt.Fprint(w, Mark, f(d))
	} else {
		fmt.Fprintf(w, " next[%#x]", h.NextHeader)
	}
}
