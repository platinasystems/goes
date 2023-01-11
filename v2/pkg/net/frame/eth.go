// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/universal"
)

// https://en.wikipedia.org/wiki/Ethernet_frame
type Eth struct {
	DMAC *universal.HardwareAddr
	SMAC *universal.HardwareAddr
	TYPE *big.Uint16
	Data []byte
}

func NewEth(data []byte) *Eth {
	eth := new(Eth)
	eth.Write(data)
	return eth

}

func (eth *Eth) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "eth:")
	fmt.Fprint(w, " ", eth.DMAC.Value())
	fmt.Fprint(w, " <- ", eth.SMAC.Value())
	switch t := eth.TYPE.Value(); t {
	case 0x0800:
		fmt.Fprint(w, ProtoMark, NewIPv4(eth.Data))
	case 0x0806:
		fmt.Fprint(w, ProtoMark, NewARP(eth.Data))
	case 0x8100, 0x88a8:
		fmt.Fprint(w, ProtoMark, NewIEEE8021Q(eth.Data))
	case 0x86dd:
		fmt.Fprint(w, ProtoMark, NewIPv6(eth.Data))
	case 0x8847, 0x8848:
		fmt.Fprint(w, ProtoMark, NewMPLS(eth.Data))
	default:
		fmt.Fprintf(w, ", type[%#x]", t)
	}
}

func (eth *Eth) IsUnicast() bool   { return (eth.DMAC[0] & 1) == 0 }
func (eth *Eth) ShouldLearn() bool { return (eth.SMAC[0] & 1) == 0 }

func (eth *Eth) DA() uint64 { return ea(eth.DMAC) }
func (eth *Eth) SA() uint64 { return ea(eth.SMAC) }

func (eth *Eth) Write(data []byte) (int, error) {
	eth.DMAC, eth.Data = universal.NewHardwareAddr(data)
	eth.SMAC, eth.Data = universal.NewHardwareAddr(eth.Data)
	eth.TYPE, eth.Data = big.NewUint16(eth.Data)
	return len(data) - len(eth.Data), nil
}

func ea(ha *universal.HardwareAddr) uint64 {
	ea := uint64(ha[0]) << 40
	ea |= uint64(ha[1]) << 32
	ea |= uint64(ha[2]) << 24
	ea |= uint64(ha[3]) << 16
	ea |= uint64(ha[4]) << 8
	ea |= uint64(ha[5])
	return ea
}
