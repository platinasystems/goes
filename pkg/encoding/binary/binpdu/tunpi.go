// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binpdu

import (
	"bytes"
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binph"
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

func (pdu TunPI) Format(w fmt.State, verb rune) {
	var h binph.TunPI
	buf := bytes.NewBuffer(pdu)
	fmt.Fprint(w, "tun")
	if _, err := h.ReadFrom(buf); err != nil {
		fmt.Fprint(w, err)
	} else if f, ok := TunPIprotos[h.Proto]; ok {
		fmt.Fprint(w, Mark, f(buf.Bytes()))
	} else {
		fmt.Fprintf(w, " proto[%#x]", h.Proto)
	}
}
