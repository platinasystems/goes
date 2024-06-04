// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binph"
)

type TCP []byte

func (pdu TCP) Format(w fmt.State, verb rune) {
	var h binph.TCP
	fmt.Fprint(w, "tcp ")
	if payload := h.PullFrom(pdu); len(payload) == len(pdu) {
		fmt.Fprint(w, ErrUnderrun)
	} else {
		fmt.Fprint(w, "port ", h.DP, " <- ", h.SP)
	}
}
