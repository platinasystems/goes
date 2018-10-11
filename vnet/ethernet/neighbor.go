// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ethernet

import (
	"github.com/platinasystems/go/elib/cpu"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"

	"errors"
	"fmt"
)

type ipNeighborFamily struct {
	m              *ip.Main
	pool           ipNeighborPool
	indexByAddress map[ipNeighborKey]uint
}

type ipNeighborMain struct {
	v *vnet.Vnet
	// Ip4/Ip6 neighbors.
	ipNeighborFamilies [ip.NFamily]ipNeighborFamily
}

func (m *ipNeighborMain) init(v *vnet.Vnet, im4, im6 *ip.Main) {
	m.v = v
	m.ipNeighborFamilies[ip.Ip4].m = im4
	m.ipNeighborFamilies[ip.Ip6].m = im6
	v.RegisterSwIfAddDelHook(m.swIfAddDel)
}

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

func (m *ipNeighborMain) AddDelIpNeighbor(im *ip.Main, n *IpNeighbor, isDel bool) (ai ip.Adj, err error) {
	ai = ip.AdjNil
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
		as     []ip.Adjacency
		prefix ip.Prefix
	)
	prefix.Address = n.Ip
	prefix.Len = 32
	if im.Family == ip.Ip6 {
		prefix.Len = 128
	}
	if ok {
		ai, as, ok = im.GetRoute(&prefix, n.Si)
		// Delete from map both of add and delete case.
		// For add case we'll re-add to indexByAddress.
		delete(nf.indexByAddress, k)
	}
	if isDel {
		if len(as) > 0 {
			if vnet.AdjDebug {
				fmt.Printf("neighbor.go call AddDelRoute to delete %v adj %v from %v\n", prefix.Address, ai.String(), n.Si.Name(m.v))
			}
			if _, err = im.AddDelRoute(&prefix, im.FibIndexForSi(n.Si), ai, isDel); err != nil {
				return
			}

			im.DelAdj(ai)
		}

		ai = ip.AdjNil
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

		if vnet.AdjDebug {
			fmt.Printf("neighbor.go call AddDelRoute to add %v adj %v to %v\n", prefix.Address, ai.String(), n.Si.Name(m.v))
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

func (m *ipNeighborMain) delKey(nf *ipNeighborFamily, k *ipNeighborKey) (err error) {
	n := IpNeighbor{
		Ip: k.Ip,
		Si: k.Si,
	}
	const isDel = true
	_, err = m.AddDelIpNeighbor(nf.m, &n, isDel)
	return
}

func (m *ipNeighborMain) swIfAddDel(v *vnet.Vnet, si vnet.Si, isDel bool) (err error) {
	if isDel {
		for fi := range m.ipNeighborFamilies {
			nf := &m.ipNeighborFamilies[fi]
			for k, _ := range nf.indexByAddress {
				if k.Si == si {
					if err = m.delKey(nf, &k); err != nil {
						return
					}
				}
			}
		}
	}
	return
}
