// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
)

// https://en.wikipedia.org/wiki/User_Datagram_Protocol
type UDP struct {
	SP  big.Uint16
	DP  big.Uint16
	Len big.Uint16
	Sum big.Uint16
}

func (udp *UDP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "udp: ", udp.DP.Value(), " <- ", udp.SP.Value())
}
