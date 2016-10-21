// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ethernet

import (
	"github.com/platinasystems/go/elib/cpu"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"

	"errors"
)

type ipNeighborFamily struct {
	pool           ipNeighborPool
	indexByAddress map[ipNeighborKey]uint
}

type ipNeighborMain struct {
	v *vnet.Vnet
	// Ip4/Ip6 neighbors.
	ipNeighborFamilies [ip.NFamily]ipNeighborFamily
}

func (m *ipNeighborMain) init(v *vnet.Vnet) { m.v = v }

type ipNeighborKey struct {
	Ip ip.Address
	Si vnet.Si
}

type IpNeighbor struct {
	Ethernet Address
	Ip       ip.Address
	Si       vnet.Si
}

type ipNeighbor struct {
	IpNeighbor
	index        uint
	lastTimeUsed cpu.Time
}

//go:generate gentemplate -d Package=ethernet -id ipNeighbor -d PoolType=ipNeighborPool -d Data=neighbors -d Type=ipNeighbor github.com/platinasystems/go/elib/pool.tmpl

var ErrDelUnknownNeighbor = errors.New("delete unknown neighbor")

func (m *ipNeighborMain) AddDelIpNeighbor(im *ip.Main, n *IpNeighbor, isDel bool) (err error) {
	nf := &m.ipNeighborFamilies[im.Family]

	var (
		k  ipNeighborKey
		i  uint
		ok bool
	)
	k.Si, k.Ip = n.Si, n.Ip
	if i, ok = nf.indexByAddress[k]; !ok {
		if isDel {
			err = ErrDelUnknownNeighbor
			return
		}
		i = nf.pool.GetIndex()
	}
	in := &nf.pool.neighbors[i]

	var (
		ai     ip.Adj
		as     []ip.Adjacency
		prefix ip.Prefix
	)
	prefix.Address = n.Ip
	prefix.Len = 32
	if im.Family == ip.Ip6 {
		prefix.Len = 128
	}
	if ok {
		ai, ok = im.GetRoute(&prefix, n.Si)
		if !ok {
			panic("get route")
		}
		as = im.GetAdj(ai)
		delete(nf.indexByAddress, k)
	}
	if isDel {
		*in = ipNeighbor{}
	} else {
		is_new_adj := len(as) == 0
		if is_new_adj {
			ai, as = im.NewAdj(1)
		}
		m.v.SetRewrite(&as[0].Rewrite, n.Si, im.RewriteNode, im.PacketType, n.Ethernet[:])
		as[0].LookupNextIndex = ip.LookupNextRewrite

		if is_new_adj {
			im.CallAdjAddHooks(ai)
		}

		if _, err = im.AddDelRoute(&prefix, im.FibIndexForSi(n.Si), ai, isDel); err != nil {
			return
		}

		// Update neighbor fields (ethernet address may change).
		in.IpNeighbor = *n
		in.index = i
		in.lastTimeUsed = cpu.TimeNow()

		if nf.indexByAddress == nil {
			nf.indexByAddress = make(map[ipNeighborKey]uint)
		}
		nf.indexByAddress[k] = i
	}

	return
}
