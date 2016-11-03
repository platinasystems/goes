// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package main

import (
	"os"

	"github.com/platinasystems/go/builtinutils"
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/coreutils"
	"github.com/platinasystems/go/diagutils"
	"github.com/platinasystems/go/diagutils/dlv"
	"github.com/platinasystems/go/fsutils"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/initutils"
	"github.com/platinasystems/go/initutils/start"
	"github.com/platinasystems/go/kutils"
	"github.com/platinasystems/go/machined"
	"github.com/platinasystems/go/machined/info/cmdline"
	"github.com/platinasystems/go/machined/info/hostname"
	"github.com/platinasystems/go/machined/info/netlink"
	"github.com/platinasystems/go/machined/info/uptime"
	"github.com/platinasystems/go/machined/info/version"
	vnetinfo "github.com/platinasystems/go/machined/info/vnet"
	"github.com/platinasystems/go/netutils"
	vnetcmd "github.com/platinasystems/go/netutils/vnet"
	"github.com/platinasystems/go/redisutils"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"
)

func main() {
	command.Plot(builtinutils.New()...)
	command.Plot(coreutils.New()...)
	command.Plot(dlv.New()...)
	command.Plot(diagutils.New()...)
	command.Plot(fsutils.New()...)
	command.Plot(initutils.New()...)
	command.Plot(kutils.New()...)
	command.Plot(machined.New())
	command.Plot(vnetcmd.New())
	command.Plot(netutils.New()...)
	command.Plot(redisutils.New()...)
	command.Sort()
	start.Hook = func() error {
		os.Setenv("REDISD", "lo eth0")
		return nil
	}
	machined.Hook = func() error {
		machined.Plot(
			cmdline.New(),
			hostname.New(),
			netlink.New(),
			uptime.New(),
			version.New(),
			vnetinfo.New(vnetinfo.Config{
				UnixInterfacesOnly: true,
				PublishAllCounters: false,
				GdbWait:            false,
				Hook:               vnetHook,
			}),
		)
		machined.Info["netlink"].Prefixes("lo.", "eth0.")
		return nil
	}
	goes.Main()
}

func vnetHook(i *vnetinfo.Info) error {
	v := i.V()

	// Base packages.
	ethernet.Init(v)
	ip4.Init(v)
	ip6.Init(v)
	pg.Init(v)   // vnet packet generator
	unix.Init(v) // tuntap/netlink

	// Device drivers: Broadcom switch + Intel 10G ethernet for punt path.
	ixge.Init(v)
	bcm.Init(v)

	plat := &platform{i: i}
	v.AddPackage("platform", plat)
	plat.DependsOn("pci-discovery")

	return nil
}
