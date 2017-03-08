// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip

import (
	"github.com/platinasystems/go/vnet"
)

type AddressStringer func(a *Address) string

type FamilyConfig struct {
	AddressStringer  AddressStringer
	Family           Family
	RewriteNode      vnet.Noder
	PacketType       vnet.PacketType
	GetRoute         func(p *Prefix, si vnet.Si) (ai Adj, ok bool)
	GetRouteFibIndex func(p *Prefix, fi FibIndex) (ai Adj, ok bool)
	AddDelRoute      func(p *Prefix, fi FibIndex, newAdj Adj, isDel bool) (oldAdj Adj, err error)
	RemapAdjacency   func(oldAdj, newAdj Adj)
}

type Main struct {
	v *vnet.Vnet
	FamilyConfig
	fibMain
	adjacencyMain
	ifAddressMain
}

func (m *Main) Init(v *vnet.Vnet) { m.adjacencyInit() }
func (m *Main) PackageInit(v *vnet.Vnet, c FamilyConfig) {
	m.v = v
	m.FamilyConfig = c
	m.ifAddressMain.init(v)
}
