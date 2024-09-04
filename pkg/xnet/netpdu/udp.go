// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type UDP []byte

func (pdu UDP) Header() (h netph.UDP, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu UDP) Data() (d []byte) {
	if len(pdu) >= netph.SizeofUDP {
		d = []byte(pdu)[netph.SizeofUDP:]
	}
	return
}

func (pdu UDP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "udp ")
	if h, err := pdu.Header(); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, "port ", h.DP, " <- ", h.SP)
	}
}
