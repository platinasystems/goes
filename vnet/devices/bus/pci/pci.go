package pci

import (
	"github.com/platinasystems/go/elib/hw/pci"
	"github.com/platinasystems/go/vnet"
)

type pciDiscover struct{ vnet.Package }

func (d *pciDiscover) Init() error { return pci.DiscoverDevices() }
func Init(v *vnet.Vnet) {
	name := "pci-discovery"
	if _, ok := v.PackageByName(name); !ok {
		v.AddPackage(name, &pciDiscover{})
	}
}
