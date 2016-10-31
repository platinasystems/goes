package bcm

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/tomahawk"
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
	m.InitPlatform(v)
	tomahawk.Init(v)
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
