// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip

import (
	"github.com/platinasystems/go/vnet"

	"strconv"
)

type Family uint8

const (
	Ip4 Family = iota
	Ip6
	NFamily
)

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

func (p *Prefix) String(m *Main) string {
	return m.AddressStringer(&p.Address) + "/" + strconv.Itoa(int(p.Len))
}
