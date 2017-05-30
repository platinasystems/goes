// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip6

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"

	"fmt"
)

var masks = compute_masks()

func compute_masks() (m [129]Address) {
	for l := uint(0); l < uint(len(m)); l++ {
		for i := uint(0); i < l; i++ {
			m[l][i/8] |= uint8(1) << (7 - (i % 8))
		}
	}
	return
}

func (a *Address) MaskLen() (len uint, ok bool) {
	len = ^uint(0)
	j, l := uint(0), uint(0)
	for i := 0; i < AddressBytes; i++ {
		m := a[i]
		if j != 0 {
			if m != 0 {
				return
			}
		} else {
			switch m {
			case 0xff:
				l += 8
			case 0xfe:
				l += 7
			case 0xfc:
				l += 6
			case 0xf8:
				l += 5
			case 0xf0:
				l += 4
			case 0xe0:
				l += 3
			case 0xc0:
				l += 2
			case 0x80:
				l += 1
			case 0:
				l += 0
			default:
				return
			}
			if m != 0xff {
				j++
			}
		}
	}
	len = l
	ok = true
	return
}

func (v *Address) MaskedString(r vnet.MaskedStringer) (s string) {
	m := r.(*Address)
	s = v.String() + "/"
	if l, ok := m.MaskLen(); ok {
		s += fmt.Sprintf("%d", l)
	} else {
		s += fmt.Sprintf("%s", m.String())
	}
	return
}

func AddressMaskForLen(l uint) Address { return masks[l] }

type Prefix struct {
	Address
	Len uint32
}

func (p *Prefix) SetLen(l uint) { p.Len = uint32(l) }
func (a *Address) toPrefix() (p Prefix) {
	p.Address = *a
	return
}

func FromIp6Prefix(i *ip.Prefix) (p Prefix) {
	copy(p.Address[:], i.Address[:AddressBytes])
	p.Len = i.Len
	return
}
func (p *Prefix) ToIpPrefix() (i ip.Prefix) {
	copy(i.Address[:], p.Address[:])
	i.Len = p.Len
	return
}
