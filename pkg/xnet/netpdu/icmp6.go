// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

var ICMP6TypeCodes = map[uint8]func(uint8) string{
	ICMP6TypeDestinationUnreachable: ICMP6UnreachableCodeName[uint8],
	ICMP6TypeTimeExceeded:           ICMP6TimeExceededCodeName[uint8],
	ICMP6TypeInvalidParameter:       ICMP6InvalidParameterCodeName[uint8],
	ICMP6TypeRouterRenumbering:      ICMP6RouterRenumberingCodeName[uint8],
}

type ICMP6 []byte

func (pdu ICMP6) Header() (h netph.ICMP6, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu ICMP6) Data() (d []byte) {
	if len(pdu) >= netph.SizeofICMP6 {
		d = []byte(pdu)[netph.SizeofICMP6:]
	}
	return
}

func (pdu ICMP6) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "icmp6 ")
	h, err := pdu.Header()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, ICMP6TypeName(h.Type), ", ")
	if f, ok := ICMP6TypeCodes[h.Type]; ok {
		fmt.Fprint(w, f(h.Code))
	} else {
		fmt.Fprint(w, h.Code)
	}
}
