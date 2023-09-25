// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
)

// https://en.wikipedia.org/wiki/IEEE_802.1Q
type IEEE8021Q struct {
	TCI  big.Uint16
	TYPE big.Uint16
}

func (q *IEEE8021Q) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ieee802_1Q:")
	fmt.Fprintf(w, " pcp %#x", q.PCP())
	fmt.Fprint(w, ", dei ", q.DEI())
	fmt.Fprintf(w, ", vid %#x", q.VID())
	switch t := q.TYPE.Value(); t {
	case 0x0800:
		fmt.Fprint(w, ProtoMark, (*IPv4)(Data(q)))
	case 0x0806:
		fmt.Fprint(w, ProtoMark, (*ARP)(Data(q)))
	case 0x8100:
		fmt.Fprint(w, ProtoMark, (*IEEE8021Q)(Data(q)))
	case 0x86dd:
		fmt.Fprint(w, ProtoMark, (*IPv6)(Data(q)))
	case 0x8847, 0x8848:
		fmt.Fprint(w, ProtoMark, (*MPLS)(Data(q)))
	default:
		fmt.Fprintf(w, ", type[%#x]", t)
	}
}

func (q *IEEE8021Q) PCP() uint8  { return uint8(q.TCI.Value() >> (1 + 12)) }
func (q *IEEE8021Q) DEI() bool   { return (q.TCI.Value() & (1 << 12)) != 0 }
func (q *IEEE8021Q) VID() uint16 { return q.TCI.Value() & ((1 << 12) - 1) }

func (q *IEEE8021Q) SetTCI(pcp uint8, dei bool, vid uint16) {
	vid &= (1 << 12) - 1
	if dei {
		vid |= 1 << 12
	}
	vid |= uint16(pcp&7) << (1 + 12)
	q.TCI.Put(vid)
}
