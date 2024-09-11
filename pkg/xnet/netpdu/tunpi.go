// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

var TunPIprotos = map[uint16]func([]byte) fmt.Formatter{
	netph.ETH_P_IP: func(data []byte) fmt.Formatter {
		return IP(data)
	},
	netph.ETH_P_IPV6: func(data []byte) fmt.Formatter {
		return IP6(data)
	},
}

type TunPI []byte

func (pdu TunPI) Header() (h netph.TunPI, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu TunPI) Data() (d []byte) {
	if len(pdu) >= netph.TunPISize {
		d = []byte(pdu)[netph.TunPISize:]
	}
	return
}

func (pdu TunPI) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "tun")
	if h, err := pdu.Header(); err != nil {
		fmt.Fprint(w, " ", err)
	} else if f, ok := TunPIprotos[h.Proto]; ok {
		fmt.Fprint(w, Mark, f(pdu.Data()))
	} else {
		fmt.Fprintf(w, " proto[%#x]", h.Proto)
	}
}
