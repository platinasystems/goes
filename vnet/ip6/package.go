package ip6

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"
)

var packageIndex uint

func Init(v *vnet.Vnet) {
	m := &Main{}
	packageIndex = v.AddPackage("ip6", m)
}

func GetMain(v *vnet.Vnet) *Main { return v.GetPackage(packageIndex).(*Main) }

func ipAddressStringer(a *ip.Address) string { return IpAddress(a).String() }

type Main struct {
	vnet.Package
	ip.Main
	nodeMain
}

func (m *Main) Init() (err error) {
	v := m.Vnet
	cf := ip.FamilyConfig{
		Family:          ip.Ip6,
		AddressStringer: ipAddressStringer,
		RewriteNode:     &m.rewriteNode,
		PacketType:      vnet.IP6,
	}
	m.Main.Init(v, cf)
	m.nodeInit(v)

	return
}
