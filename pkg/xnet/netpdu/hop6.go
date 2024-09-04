// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

var HOP6Types = map[uint8]func([]byte) fmt.Formatter{
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

type HOP6 []byte

func (pdu HOP6) Header() (h netph.HOP6, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu HOP6) Data() (d []byte) {
	if len(pdu) >= netph.SizeofHOP6 {
		d = []byte(pdu)[netph.SizeofHOP6:]
	}
	return
}

func (pdu HOP6) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "hop6 ")
	if h, err := pdu.Header(); err != nil {
		fmt.Fprint(w, err)
	} else if s, ok := map[uint8]string{
		0:   "hop-by-hop",
		43:  "routing",
		44:  "fragment",
		50:  "ESP",
		51:  "AH",
		60:  "destination options",
		135: "mobility",
		139: "HIP",
		140: "shim",
	}[h.Type]; ok {
		fmt.Fprintf(w, "%s[%d]", s, h.Len)
		if h.Type != 59 {
			fmt.Fprint(w, Mark, HOP6(pdu.Data()))
		}
	} else if f, ok := HOP6Types[h.Type]; ok {
		fmt.Fprint(w, Mark, f(pdu.Data()))
	} else {
		fmt.Fprintf(w, "type[%#x]", h.Type)
	}
}
