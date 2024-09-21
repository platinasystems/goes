// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type MPLS_UC []byte
type MPLS_MC []byte
type MPLS []byte

func (pdu MPLS_UC) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "mpls-uc ", MPLS(pdu))
}

func (pdu MPLS_MC) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "mpls-mc ", MPLS(pdu))
}

func (pdu MPLS) Format(w fmt.State, verb rune) {
	h, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprintf(w, "label %#x, tc %#x, ttl %d", h.Label(), h.TC(), h.TTL())
	if len(d) < 2 {
		return
	}
	if !h.IsBOS() {
		fmt.Fprint(w, Mark, MPLS(d))
		return
	}
	t := uint16(d[0])<<8 | uint16(d[1])
	if f, ok := MPLStypes[t]; ok {
		fmt.Fprint(w, f(d))
	} else {
		fmt.Fprintf(w, ", type %#x", t)
	}
}

func (pdu MPLS) Parse() (netph.MPLS, []byte, error) {
	return netph.Parse[netph.MPLS](pdu)
}

var MPLStypes = map[uint16]func([]byte) fmt.Formatter{
	netph.ETH_P_IP: func(data []byte) fmt.Formatter {
		return IP(data)
	},
	netph.ETH_P_IPV6: func(data []byte) fmt.Formatter {
		return IP6(data)
	},
}
