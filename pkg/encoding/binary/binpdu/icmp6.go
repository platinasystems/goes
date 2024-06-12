// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binpdu

import (
	"bytes"
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binph"
)

var ICMP6TypeCodes = map[uint8]func(uint8) string{
	ICMP6TypeDestinationUnreachable: ICMP6UnreachableCodeName[uint8],
	ICMP6TypeTimeExceeded:           ICMP6TimeExceededCodeName[uint8],
	ICMP6TypeInvalidParameter:       ICMP6InvalidParameterCodeName[uint8],
	ICMP6TypeRouterRenumbering:      ICMP6RouterRenumberingCodeName[uint8],
}

type ICMP6 []byte

func (pdu ICMP6) Format(w fmt.State, verb rune) {
	var h binph.ICMP6
	buf := bytes.NewBuffer(pdu)
	fmt.Fprint(w, "icmp6 ")
	if _, err := h.ReadFrom(buf); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, ICMP6TypeName(h.Type), ", ")
		if f, ok := ICMP6TypeCodes[h.Type]; ok {
			fmt.Fprint(w, f(h.Code))
		} else {
			fmt.Fprint(w, h.Code)
		}
	}
}
