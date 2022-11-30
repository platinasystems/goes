// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
)

// https://en.wikipedia.org/wiki/Ethernet_frame
type Eth []byte

func (eth Eth) DMAC() net.HardwareAddr { return net.HardwareAddr(eth[:6]) }
func (eth Eth) SMAC() net.HardwareAddr { return net.HardwareAddr(eth[6 : 6+6]) }
func (eth Eth) TYPE() big.Uint16       { return big.Uint16(eth[12 : 12+2]) }
func (eth Eth) Data() []byte           { return []byte(eth[14:]) }

func (eth Eth) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "eth:")
	fmt.Fprint(w, " ", eth.SMAC())
	fmt.Fprint(w, " -> ", eth.DMAC())
	switch eth.TYPE().Value() {
	case 0x0800:
		fmt.Fprint(w, ProtoMark, IP(eth.Data()))
	case 0x0806:
		fmt.Fprint(w, ProtoMark, ARP(eth.Data()))
	case 0x8100:
		fmt.Fprint(w, ProtoMark, IEEE802_1Q(eth.Data()))
	case 0x86dd:
		fmt.Fprint(w, ProtoMark, IP(eth.Data()))
	case 0x88a8:
		fmt.Fprint(w, ProtoMark, IEEE802_1AD(eth.Data()))
	case 0x8847:
		fmt.Fprint(w, ProtoMark, MPLS(eth.Data()))
	case 0x8848:
		fmt.Fprint(w, ProtoMark, MMPLS(eth.Data()))
	default:
		fmt.Fprintf(w, ", type[%#x]", eth.TYPE())
	}
}

func (eth Eth) IsUnicast() bool   { return (eth.DMAC()[0] & 1) == 0 }
func (eth Eth) ShouldLearn() bool { return (eth.SMAC()[0] & 1) == 0 }

func (eth Eth) DA() uint64 { return EA(eth.DMAC()) }
func (eth Eth) SA() uint64 { return EA(eth.SMAC()) }

func EA(ha net.HardwareAddr) uint64 {
	_ = ha[5]
	ea := uint64(ha[0]) << 40
	ea |= uint64(ha[1]) << 32
	ea |= uint64(ha[2]) << 24
	ea |= uint64(ha[3]) << 16
	ea |= uint64(ha[4]) << 8
	ea |= uint64(ha[5])
	return ea
}
