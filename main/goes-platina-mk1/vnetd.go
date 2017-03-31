// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/internal/goes/cmd/vnetd"
	"github.com/platinasystems/go/internal/prog"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/firmware"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"
)

func init() {
	vnetd.UnixInterfacesOnly = true
	vnetd.GdbWait = gdbwait
	vnetd.Hook = func(i *vnetd.Info, v *vnet.Vnet) error {
		err := firmware.Extract(prog.Name())
		if err != nil {
			return err
		}

		// Base packages.
		ethernet.Init(v)
		ip4.Init(v)
		ip6.Init(v)
		pg.Init(v) // vnet packet generator
		unix.Init(v)

		// Device drivers.
		ixge.Init(v)
		fe1.Init(v)

		plat := &platform{i: i}
		v.AddPackage("platform", plat)
		plat.DependsOn("pci-discovery")

		// Need FE1 init/port init to complete before default
		// fib/adjacencies can be installed.
		plat.DependedOnBy("ip4")
		plat.DependedOnBy("ip6")

		return nil
	}
}
