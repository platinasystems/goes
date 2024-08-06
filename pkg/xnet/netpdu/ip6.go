// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"bytes"
	"fmt"
	"net"

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

func (pdu IP6) Parse() (header netph.IP6, data []byte, err error) {
	buf := bytes.NewBuffer(pdu)
	if _, err = header.ReadFrom(buf); err == nil {
		data = buf.Bytes()
	}
	return
}

func (pdu IP6) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ip6 ")
	header, data, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, net.IP(header.DA[:]), " <- ", net.IP(header.SA[:]))
	if f, ok := IP6NextHeaders[header.NextHeader]; ok {
		fmt.Fprint(w, Mark, f(data))
	} else {
		fmt.Fprintf(w, " next[%#x]", header.NextHeader)
	}
}
