// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
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
	switch h.Proto {
	case netph.TUN_P_IP:
		fmt.Fprint(w, Mark, IP(d))
	case netph.TUN_P_IP6:
		fmt.Fprint(w, Mark, IP6(d))
	default:
		fmt.Fprintf(w, " proto %#x", h.Proto)
	}
}

func (pdu TunPI) Parse() (netph.TunPI, []byte, error) {
	return netph.Parse[netph.TunPI](pdu)
}
