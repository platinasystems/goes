// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
)

// https://en.wikipedia.org/wiki/Multiprotocol_Label_Switching
type MPLS struct{ big.Uint32 }

func (mpls *MPLS) BOS() *BOS   { return (*BOS)(Data(mpls)) }
func (mpls *MPLS) MPLS() *MPLS { return (*MPLS)(Data(mpls)) }

func (mpls *MPLS) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "mpls:")
	fmt.Fprintf(w, " label %#x", mpls.Label())
	fmt.Fprintf(w, ", tc %#x", mpls.TC())
	fmt.Fprintf(w, ", ttl %d", mpls.TTL())
	if mpls.IsBOS() {
		fmt.Fprint(w, ProtoMark, mpls.BOS())
	} else {
		fmt.Fprint(w, ProtoMark, mpls.MPLS())
	}
}

func (mpls *MPLS) Label() uint32 { return mpls.Value() >> (3 + 1 + 8) }
func (mpls *MPLS) TC() uint8     { return uint8(mpls.Value()>>9) & 7 }
func (mpls *MPLS) IsBOS() bool   { return (mpls.Value() & (1 << 8)) != 0 }
func (mpls *MPLS) TTL() uint8    { return uint8(mpls.Value()) }

func (mpls *MPLS) Set(label uint32, tc uint8, isbos bool, ttl uint8) {
	h := uint32(ttl)
	if isbos {
		h |= (1 << 8)
	}
	h |= uint32(tc & 7)
	h |= (label & ((1 << 20) - 1)) << (3 + 1 + 8)
	mpls.Put(h)
}

// Bottom Of [MPLS] Stack
type BOS struct{ big.Uint32 }

func (bos *BOS) Format(w fmt.State, verb rune) {
	switch t := bos.Value(); t {
	case 0x0800:
		fmt.Fprint(w, ProtoMark, (*IPv4)(Data(bos)))
	case 0x0806:
		fmt.Fprint(w, ProtoMark, (*ARP)(Data(bos)))
	case 0x8100, 0x88a8:
		fmt.Fprint(w, ProtoMark, (*IEEE8021Q)(Data(bos)))
	case 0x86dd:
		fmt.Fprint(w, ProtoMark, (*IPv6)(Data(bos)))
	default:
		fmt.Fprintf(w, "type[%#x]", t)
	}
}
