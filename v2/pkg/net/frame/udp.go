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
	SP   *big.Uint16
	DP   *big.Uint16
	Len  *big.Uint16
	Sum  *big.Uint16
	Data []byte
}

func NewUDP(data []byte) *UDP {
	udp := new(UDP)
	udp.Write(data)
	return udp
}

func (udp *UDP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "udp[", udp.Len.Value(), "]: ",
		udp.DP.Value(), " <- ", udp.SP.Value())
}

func (udp *UDP) Write(data []byte) (int, error) {
	udp.SP, udp.Data = big.NewUint16(data)
	udp.DP, udp.Data = big.NewUint16(udp.Data)
	udp.Len, udp.Data = big.NewUint16(udp.Data)
	udp.Sum, udp.Data = big.NewUint16(udp.Data)
	return len(data) - len(udp.Data), nil
}
