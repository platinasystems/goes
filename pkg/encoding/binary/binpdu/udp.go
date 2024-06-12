// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binpdu

import (
	"bytes"
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binph"
)

type UDP []byte

func (pdu UDP) Format(w fmt.State, verb rune) {
	var h binph.UDP
	buf := bytes.NewBuffer(pdu)
	fmt.Fprint(w, "udp ")
	if _, err := h.ReadFrom(buf); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, "port ", h.DP, " <- ", h.SP)
	}
}
