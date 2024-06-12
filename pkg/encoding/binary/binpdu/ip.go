// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binpdu

import (
	"bytes"
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binph"
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

func (pdu IP) Format(w fmt.State, verb rune) {
	var h binph.IP
	buf := bytes.NewBuffer(pdu)
	fmt.Fprint(w, "ip ")
	if _, err := h.ReadFrom(buf); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, net.IP(h.DA[:]), " <- ", net.IP(h.SA[:]))
		i := h.IHL() * 4
		if f, ok := IPprotocols[h.Protocol]; ok && i < len(pdu) {
			fmt.Fprint(w, Mark, f(buf.Bytes()))
		} else {
			fmt.Fprintf(w, ", protocol[%#x]", h.Protocol)
		}
	}
}
