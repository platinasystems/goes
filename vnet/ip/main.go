package ip

import (
	"github.com/platinasystems/go/vnet"
)

type AddressStringer func(a *Address) string

type FamilyConfig struct {
	AddressStringer AddressStringer
	Family          Family
	RewriteNode     vnet.Noder
	PacketType      vnet.PacketType
	GetRoute        func(p *Prefix, si vnet.Si) (ai Adj, ok bool)
	AddDelRoute     func(p *Prefix, si vnet.Si, newAdj Adj, isDel bool) (oldAdj Adj, ok bool)
}

type Main struct {
	v *vnet.Vnet
	FamilyConfig
	fibMain
	adjacencyMain
	ifAddressMain
}

func (m *Main) Init(v *vnet.Vnet, c FamilyConfig) {
	m.v = v
	m.FamilyConfig = c
	m.adjacencyInit()
	m.ifAddressMain.init(v)
}
