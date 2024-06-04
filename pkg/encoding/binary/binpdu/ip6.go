// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binpdu

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binph"
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

func (pdu IP6) Format(w fmt.State, verb rune) {
	var h binph.IP6
	fmt.Fprint(w, "ip6 ")
	if payload := h.PullFrom(pdu); len(payload) == len(pdu) {
		fmt.Fprint(w, ErrUnderrun)
	} else {
		fmt.Fprint(w, net.IP(h.DA[:]), " <- ", net.IP(h.SA[:]))
		if f, ok := IP6NextHeaders[h.NextHeader]; ok {
			fmt.Fprint(w, Mark, f(payload))
		} else {
			fmt.Fprintf(w, " next[%#x]", h.NextHeader)
		}
	}
}
