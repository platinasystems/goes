// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
)

// https://en.wikipedia.org/wiki/Transmission_Control_Protocol
type TCP struct {
	SP    big.Uint16
	DP    big.Uint16
	Seq   big.Uint32
	Ack   big.Uint32
	Flags big.Uint32
	Sum   big.Uint16
	UP    big.Uint16
}

func (tcp *TCP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "tcp: ", tcp.DP.Value(), " <- ", tcp.SP.Value())
}

func (tcp *TCP) DataOffset() uint8 { return uint8(tcp.Flags[0] >> 4) }

func (tcp *TCP) NS() bool  { return (tcp.Flags.Value() & (1 << 24)) != 0 }
func (tcp *TCP) CWR() bool { return (tcp.Flags.Value() & (1 << 23)) != 0 }
func (tcp *TCP) ECE() bool { return (tcp.Flags.Value() & (1 << 22)) != 0 }
func (tcp *TCP) URG() bool { return (tcp.Flags.Value() & (1 << 21)) != 0 }
func (tcp *TCP) ACK() bool { return (tcp.Flags.Value() & (1 << 20)) != 0 }
func (tcp *TCP) PSH() bool { return (tcp.Flags.Value() & (1 << 19)) != 0 }
func (tcp *TCP) RST() bool { return (tcp.Flags.Value() & (1 << 18)) != 0 }
func (tcp *TCP) SYN() bool { return (tcp.Flags.Value() & (1 << 17)) != 0 }
func (tcp *TCP) FIN() bool { return (tcp.Flags.Value() & (1 << 16)) != 0 }

func (tcp *TCP) Window() uint16 { return uint16(tcp.Flags.Value()) }
