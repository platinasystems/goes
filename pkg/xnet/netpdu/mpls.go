// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

var MPLStypes = map[uint16]func([]byte) fmt.Formatter{
	ETH_P_IP: func(data []byte) fmt.Formatter {
		return IP(data)
	},
	ETH_P_IPV6: func(data []byte) fmt.Formatter {
		return IP6(data)
	},
}

type MPLS_UC []byte
type MPLS_MC []byte

func (pdu MPLS_UC) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "mpls-uc: ")
	mplsFormat(w, verb, pdu)
}

func (pdu MPLS_MC) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "mpls-uc: ")
	mplsFormat(w, verb, pdu)
}

func mplsFormat[PDU MPLS_UC | MPLS_MC](w fmt.State, verb rune, pdu PDU) {
	var h netph.MPLS
	buf := bytes.NewBuffer(pdu)
	if _, err := h.ReadFrom(buf); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprintf(w, "label[%#x] tc[%#x] ttl[%d]",
			h.Label(), h.TC(), h.TTL())
		payload := buf.Bytes()
		t := binary.BigEndian.Uint16(payload)
		if !h.IsBOS() {
			fmt.Fprint(w, Mark, payload)
		} else if f, ok := MPLStypes[t]; ok {
			fmt.Fprint(w, Mark, f(payload))
		} else {
			fmt.Fprintf(w, " type[%#x]", t)
		}
	}
}
