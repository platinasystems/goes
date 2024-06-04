// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binph"
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

func (pdu HOP6) Format(w fmt.State, verb rune) {
	var h binph.HOP6
	fmt.Fprint(w, "hop6 ")
	if payload := h.PullFrom(pdu); len(payload) == len(pdu) {
		fmt.Fprint(w, ErrUnderrun)
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
			fmt.Fprint(w, Mark, payload)
		}
	} else if f, ok := HOP6Types[h.Type]; ok {
		fmt.Fprint(w, Mark, f(payload))
	} else {
		fmt.Fprintf(w, "type[%#x]", h.Type)
	}
}
