// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

// https://en.wikipedia.org/wiki/IEEE_802.1Q
type IEEE8021 struct {
	TCI  uint16
	Type uint16
}

func (p *IEEE8021) PCP() uint8  { return uint8(p.TCI >> (1 + 12)) }
func (p *IEEE8021) DEI() bool   { return (p.TCI & (1 << 12)) != 0 }
func (p *IEEE8021) VID() uint16 { return p.TCI & ((1 << 12) - 1) }

func ConstructTCI(pcp uint8, dei bool, vid uint16) uint16 {
	vid &= (1 << 12) - 1
	if dei {
		vid |= 1 << 12
	}
	return vid | (uint16(pcp&7) << (1 + 12))
}
