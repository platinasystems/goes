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

func (pdu IP) Parse() (header netph.IP, data []byte, err error) {
	buf := bytes.NewBuffer(pdu)
	if _, err = header.ReadFrom(buf); err == nil {
		i := header.IHL() << 2
		if n := len(pdu); i > n {
			i = n
		}
		data = []byte(pdu[i:])
	}
	return
}

func (pdu IP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ip ")
	header, data, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, net.IP(header.DA[:]), " <- ", net.IP(header.SA[:]))
	if f, ok := IPprotocols[header.Protocol]; ok {
		fmt.Fprint(w, Mark, f(data))
	} else {
		fmt.Fprintf(w, ", protocol[%#x]", header.Protocol)
	}
}
