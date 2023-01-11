// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
)

// https://en.wikipedia.org/wiki/Multiprotocol_Label_Switching
type MPLS struct {
	Header *big.Uint32
	Data   []byte
}

func NewMPLS(data []byte) *MPLS {
	mpls := new(MPLS)
	mpls.Write(data)
	return mpls
}

func (mpls *MPLS) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "mpls:")
	fmt.Fprintf(w, " label %#x", mpls.Label())
	fmt.Fprintf(w, ", tc %#x", mpls.TC())
	fmt.Fprintf(w, ", ttl %d", mpls.TTL())
	if mpls.BOS() {
		mt, data := big.NewUint16(mpls.Data)
		switch t := mt.Value(); t {
		case 0x0800:
			fmt.Fprint(w, ProtoMark, NewIPv4(data))
		case 0x0806:
			fmt.Fprint(w, ProtoMark, NewARP(data))
		case 0x8100, 0x88a8:
			fmt.Fprint(w, ProtoMark, NewIEEE8021Q(data))
		case 0x86dd:
			fmt.Fprint(w, ProtoMark, NewIPv6(data))
		default:
			fmt.Fprintf(w, ", type[%#x]", t)
		}
	} else {
		fmt.Fprint(w, ProtoMark, NewMPLS(mpls.Data))
	}
}

func (mpls *MPLS) Label() uint32 { return mpls.Header.Value() >> (3 + 1 + 8) }
func (mpls *MPLS) TC() uint8     { return uint8(mpls.Header.Value()>>9) & 7 }
func (mpls *MPLS) BOS() bool     { return (mpls.Header.Value() & (1 << 8)) != 0 }
func (mpls *MPLS) TTL() uint8    { return uint8(mpls.Header.Value()) }

func (mpls *MPLS) Set(label uint32, tc uint8, bos bool, ttl uint8) {
	h := uint32(ttl)
	if bos {
		h |= (1 << 8)
	}
	h |= uint32(tc & 7)
	h |= (label & ((1 << 20) - 1)) << (3 + 1 + 8)
	mpls.Header.Put(h)
}

func (mpls *MPLS) Write(data []byte) (int, error) {
	mpls.Header, mpls.Data = big.NewUint32(data)
	return len(data) - len(mpls.Data), nil
}
