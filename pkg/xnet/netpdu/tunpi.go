// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"bytes"
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

var TunPIprotos = map[uint16]func([]byte) fmt.Formatter{
	ETH_P_IP: func(data []byte) fmt.Formatter {
		return IP(data)
	},
	ETH_P_IPV6: func(data []byte) fmt.Formatter {
		return IP6(data)
	},
}

type TunPI []byte

func (pdu TunPI) Parse() (header netph.TunPI, data []byte, err error) {
	buf := bytes.NewBuffer(pdu)
	if _, err = header.ReadFrom(buf); err == nil {
		data = buf.Bytes()
	}
	return
}

func (pdu TunPI) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "tun")
	if header, data, err := pdu.Parse(); err != nil {
		fmt.Fprint(w, " ", err)
	} else if f, ok := TunPIprotos[header.Proto]; ok {
		fmt.Fprint(w, Mark, f(data))
	} else {
		fmt.Fprintf(w, " proto[%#x]", header.Proto)
	}
}
