// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/omni"
)

// https://en.wikipedia.org/wiki/Ethernet_frame
type ETH struct {
	DMAC omni.HardwareAddr
	SMAC omni.HardwareAddr
	TYPE big.Uint16
}

func (eth *ETH) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "eth:")
	fmt.Fprint(w, " ", eth.DMAC.Value())
	fmt.Fprint(w, " <- ", eth.SMAC.Value())
	switch t := eth.TYPE.Value(); t {
	case 0x0800:
		fmt.Fprint(w, ProtoMark, (*IPv4)(Data(eth)))
	case 0x0806:
		fmt.Fprint(w, ProtoMark, (*ARP)(Data(eth)))
	case 0x8100, 0x88a8:
		fmt.Fprint(w, ProtoMark, (*IEEE8021Q)(Data(eth)))
	case 0x86dd:
		fmt.Fprint(w, ProtoMark, (*IPv6)(Data(eth)))
	case 0x8847, 0x8848:
		fmt.Fprint(w, ProtoMark, (*MPLS)(Data(eth)))
	default:
		fmt.Fprintf(w, ", type[%#x]", t)
	}
}

func (eth *ETH) IsUnicast() bool   { return (eth.DMAC[0] & 1) == 0 }
func (eth *ETH) ShouldLearn() bool { return (eth.SMAC[0] & 1) == 0 }

func (eth *ETH) DA() uint64 { return ea(eth.DMAC.Value()) }
func (eth *ETH) SA() uint64 { return ea(eth.SMAC.Value()) }

func ea(ha net.HardwareAddr) uint64 {
	ea := uint64(ha[0]) << 40
	ea |= uint64(ha[1]) << 32
	ea |= uint64(ha[2]) << 24
	ea |= uint64(ha[3]) << 16
	ea |= uint64(ha[4]) << 8
	ea |= uint64(ha[5])
	return ea
}
