// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binph"
)

var ICMPTypeCodes = map[uint8]func(uint8) string{
	ICMPTypeDestinationUnreachable: ICMPUnreachableCodeName[uint8],
	ICMPTypeRedirectMessage:        ICMPRedirectCodeName[uint8],
	ICMPTypeTimeExceeded:           ICMPTimeExceededCodeName[uint8],
	ICMPTypeInvalidParamter:        ICMPInvalidParameterCodeName[uint8],
	ICMPTypeExtendedEchoReply:      ICMPExtendedEchoReplyCodeName[uint8],
}

type ICMP []byte

func (pdu ICMP) Format(w fmt.State, verb rune) {
	var h binph.ICMP
	fmt.Fprint(w, "icmp ")
	if payload := h.PullFrom(pdu); len(payload) == len(pdu) {
		fmt.Fprint(w, ErrUnderrun)
	} else if typename := ICMPTypeName(h.Type); len(typename) == 0 {
		typename = "unknown"
	} else {
		fmt.Fprint(w, typename, " ")
		if f, ok := ICMPTypeCodes[h.Type]; ok {
			fmt.Fprint(w, f(h.Code))
		} else {
			fmt.Fprint(w, h.Code)
		}
	}
}
