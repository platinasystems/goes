// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bcm

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/tomahawk"
)

type Switch m.Switch

const (
	PhyInterfaceInvalid = m.PhyInterfaceInvalid
	// KR[124] backplane interface.
	PhyInterfaceKR = m.PhyInterfaceKR
	PhyInterfaceCR = m.PhyInterfaceCR
	// Serial interface to optics: SGMII, SFI/XFI, MLD
	PhyInterfaceOptics = m.PhyInterfaceOptics
)

type PortConfig m.PortConfig
type PhyConfig m.PhyConfig
type SwitchConfig struct {
	Phys               []PhyConfig
	Ports              []PortConfig
	MMUPipeByPortBlock []uint8
}

type Platform m.Platform

func Init(v *vnet.Vnet) {
	p := &m.Platform{}
	p.InitPlatform(v)
	tomahawk.RegisterDeviceIDs(v)
}
func GetPlatform(v *vnet.Vnet) *Platform { return (*Platform)(m.GetPlatform(v)) }

func (cf *SwitchConfig) Configure(v *vnet.Vnet, s Switch) {
	c := &s.GetSwitchCommon().SwitchConfig
	for i := range cf.Phys {
		c.Phys = append(c.Phys, m.PhyConfig(cf.Phys[i]))
	}
	for i := range cf.Ports {
		c.Ports = append(c.Ports, m.PortConfig(cf.Ports[i]))
	}
	c.MMUPipeByPortBlock = cf.MMUPipeByPortBlock
	s.PortInit(v)
}
