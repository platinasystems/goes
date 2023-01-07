// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"net"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
)

type EthData struct {
	*Eth
	Data []byte
}

func NewEthData(data []byte) EthData {
	eth := (*Eth)(unsafe.Pointer(&data[0]))
	return EthData{eth, data[unsafe.Sizeof(*eth):]}
}

func (eth EthData) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "eth:")
	fmt.Fprint(w, " ", eth.SMAC())
	fmt.Fprint(w, " -> ", eth.DMAC())
	switch t := eth.TYPE.Value(); t {
	case 0x0800:
		fmt.Fprint(w, ProtoMark, NewIPv4Data(eth.Data))
	case 0x0806:
		fmt.Fprint(w, ProtoMark, NewARPData(eth.Data))
	case 0x8100:
		fmt.Fprint(w, ProtoMark, NewIEEE8021QData(eth.Data))
	case 0x86dd:
		fmt.Fprint(w, ProtoMark, NewIPv6Data(eth.Data))
	case 0x88a8:
		fmt.Fprint(w, ProtoMark, NewIEEE8021ADData(eth.Data))
	case 0x8847:
		fmt.Fprint(w, ProtoMark, NewMPLSData(eth.Data))
	case 0x8848:
		fmt.Fprint(w, ProtoMark, NewMMPLSData(eth.Data))
	default:
		fmt.Fprintf(w, ", type[%#x]", t)
	}
}

// https://en.wikipedia.org/wiki/Ethernet_frame
type Eth struct {
	dmac [6]byte
	smac [6]byte
	TYPE big.Uint16
}

func (eth *Eth) DMAC() net.HardwareAddr {
	return net.HardwareAddr(eth.dmac[:])
}

func (eth *Eth) SMAC() net.HardwareAddr {
	return net.HardwareAddr(eth.smac[:])
}

func (eth *Eth) IsUnicast() bool   { return (eth.dmac[0] & 1) == 0 }
func (eth *Eth) ShouldLearn() bool { return (eth.smac[0] & 1) == 0 }

func (eth *Eth) DA() uint64 { return ea(&eth.dmac) }
func (eth *Eth) SA() uint64 { return ea(&eth.smac) }

func ea(a *[6]byte) uint64 {
	ea := uint64(a[0]) << 40
	ea |= uint64(a[1]) << 32
	ea |= uint64(a[2]) << 24
	ea |= uint64(a[3]) << 16
	ea |= uint64(a[4]) << 8
	ea |= uint64(a[5])
	return ea
}
