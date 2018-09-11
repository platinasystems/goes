// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ethernet

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"
)

var packageIndex uint

type Main struct {
	vnet.Package
	ipNeighborMain
	nodeMain
	pgMain
	m4, m6   *ip.Main
	layerMap map[Type]vnet.Layer
}

func RegisterLayer(v *vnet.Vnet, t Type, l vnet.Layer) {
	m := GetMain(v)
	if m.layerMap == nil {
		m.layerMap = make(map[Type]vnet.Layer)
	}
	m.layerMap[t] = l
}
func UnregisterLayer(v *vnet.Vnet, t Type) (ok bool) {
	m := GetMain(v)
	_, ok = m.layerMap[t]
	delete(m.layerMap, t)
	return
}

func Init(v *vnet.Vnet, m4, m6 *ip.Main) {
	m := &Main{}
	m.m4, m.m6 = m4, m6
	packageIndex = v.AddPackage("ethernet", m)
	m.DependsOn("pg")
}

func GetMain(v *vnet.Vnet) *Main { return v.GetPackage(packageIndex).(*Main) }

func (m *Main) Init() (err error) {
	v := m.Vnet
	m.ipNeighborMain.init(v, m.m4, m.m6)
	m.nodeInit(v)
	m.pgMain.pgInit(v)
	m.cliInit(v)
	return
}
