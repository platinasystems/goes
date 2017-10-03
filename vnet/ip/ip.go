// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/vnet"

	"strconv"
)

type Family uint8

const (
	Ip4 Family = iota
	Ip6
	NFamily
)

func (x Family) String() string {
	t := [...]string{
		Ip4: "ip4",
		Ip6: "ip6",
	}
	return elib.StringerHex(t[:], int(x))
}

// Generic ip4/ip6 address: big enough for either.
type Address [16]uint8

// Address 32-bit chunk in network byte order.
func (a *Address) AsUint32(i uint) vnet.Uint32 {
	return vnet.Uint32(a[4*i+3]) | vnet.Uint32(a[4*i+2])<<8 | vnet.Uint32(a[4*i+1])<<16 | vnet.Uint32(a[4*i+0])<<24
}

func (a *Address) Add(x uint64)          { vnet.ByteAdd(a[:], x) }
func AddressUint64(x uint64) (a Address) { a.Add(x); return }
func AddressUint32(x uint32) (a Address) { vnet.ByteAdd(a[:4], uint64(x)); return }

type Prefix struct {
	Address
	Len uint32
}

// NewPrefix returns prefix with given length and address.
// Address will be masked so that only length bits may be non-zero.
func NewPrefix(l uint32, a []byte) (p Prefix) {
	p.Len = l
	i, n_left := 0, int(l)
	for n_left > 0 {
		mask := byte(0xff)
		if n_left < 8 {
			mask = (1<<uint(n_left) - 1) << uint(8-n_left)
		}
		p.Address[i] = a[i] & mask
		i++
		n_left -= 8
	}
	return
}

// Returns whether p matches q and is more specific.
func (p *Prefix) IsMoreSpecific(q *Prefix) (ok bool) {
	if p.Len <= q.Len {
		return
	}
	// Compare masked bits up to less specific length.
	i, n_left := 0, int(q.Len)
	for n_left > 0 {
		mask := byte(0xff)
		if n_left < 8 {
			mask = (1<<uint(n_left) - 1) << uint(8-n_left)
		}
		ok = p.Address[i]&mask == q.Address[i]&mask
		if !ok {
			return
		}
		i++
		n_left -= 8
	}
	return
}

func (p *Prefix) String(m *Main) string {
	return m.AddressStringer(&p.Address) + "/" + strconv.Itoa(int(p.Len))
}
