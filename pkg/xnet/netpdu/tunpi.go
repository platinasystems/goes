// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type TunPI []byte

func (pdu TunPI) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "tun")
	h, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, " ", err)
		return
	}
	if f, ok := TunPIprotos[h.Proto]; ok {
		fmt.Fprint(w, Mark, f(d))
	} else {
		fmt.Fprintf(w, " proto %#x", h.Proto)
	}
}

func (pdu TunPI) Parse() (netph.TunPI, []byte, error) {
	return netph.Parse[netph.TunPI](pdu)
}

var TunPIprotos = map[uint16]func([]byte) fmt.Formatter{
	netph.ETH_P_IP: func(d []byte) fmt.Formatter {
		return IP(d)
	},
	netph.ETH_P_IPV6: func(d []byte) fmt.Formatter {
		return IP6(d)
	},
}
