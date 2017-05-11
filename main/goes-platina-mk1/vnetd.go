// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/fe1"
	"github.com/platinasystems/go/goes/cmd/vnetd"
	"github.com/platinasystems/go/internal/sriovs"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"
)

func init() {
	vnetd.Hook = func(i *vnetd.Info, v *vnet.Vnet) error {
		fns, err := sriovs.NumvfsFns()
		have_numvfs := err == nil && len(fns) > 0

		vnetd.UnixInterfacesOnly = !have_numvfs
		vnetd.GdbWait = gdbwait

		// Base packages.
		ethernet.Init(v)
		ip4.Init(v)
		ip6.Init(v)
		pg.Init(v) // vnet packet generator
		unix.Init(v)

		// Device drivers.
		fe1.Init(v)
		if !have_numvfs {
			ixge.Init(v)
		} else if err = newSriovs(); err != nil {
			return err
		}

		plat := &platform{Hook: i.Init}
		v.AddPackage("platform", plat)
		plat.DependsOn("pci-discovery")

		// Need FE1 init/port init to complete before default
		// fib/adjacencies can be installed.
		plat.DependedOnBy("ip4")
		plat.DependedOnBy("ip6")

		return nil
	}
}
