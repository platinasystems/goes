// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pci

import (
	"github.com/platinasystems/go/elib/hw/pci"
	"github.com/platinasystems/go/vnet"
)

type pciDiscover struct{ vnet.Package }

func (d *pciDiscover) Init() error { return pci.DiscoverDevices(pci.DefaultBus) }
func (d *pciDiscover) Exit() error { return pci.CloseDiscoveredDevices(pci.DefaultBus) }

func Init(v *vnet.Vnet) {
	name := "pci-discovery"
	if _, ok := v.PackageByName(name); !ok {
		v.AddPackage(name, &pciDiscover{})
	}
}
