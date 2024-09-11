// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import "unsafe"

// https://en.wikipedia.org/wiki/Transmission_Control_Protocol
type TCP struct {
	SP    uint16
	DP    uint16
	Seq   uint32
	Ack   uint32
	Flags uint32
	Sum   uint16
	UP    uint16
}

const TCPSumIndex = 16
const TCPSize = int(unsafe.Sizeof(TCP{}))

func (p *TCP) DataOffset() uint8 { return uint8(p.Flags >> 28) }

func (p *TCP) NS() bool  { return (p.Flags & (1 << 24)) != 0 }
func (p *TCP) CWR() bool { return (p.Flags & (1 << 23)) != 0 }
func (p *TCP) ECE() bool { return (p.Flags & (1 << 22)) != 0 }
func (p *TCP) URG() bool { return (p.Flags & (1 << 21)) != 0 }
func (p *TCP) ACK() bool { return (p.Flags & (1 << 20)) != 0 }
func (p *TCP) PSH() bool { return (p.Flags & (1 << 19)) != 0 }
func (p *TCP) RST() bool { return (p.Flags & (1 << 18)) != 0 }
func (p *TCP) SYN() bool { return (p.Flags & (1 << 17)) != 0 }
func (p *TCP) FIN() bool { return (p.Flags & (1 << 16)) != 0 }

func (p *TCP) Window() uint16 { return uint16(p.Flags) }
