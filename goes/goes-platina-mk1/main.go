// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package main

import (
	"os"

	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/commands/builtin"
	"github.com/platinasystems/go/commands/core"
	"github.com/platinasystems/go/commands/dlv"
	"github.com/platinasystems/go/commands/fs"
	"github.com/platinasystems/go/commands/kernel"
	"github.com/platinasystems/go/commands/machine"
	"github.com/platinasystems/go/commands/machine/machined"
	"github.com/platinasystems/go/commands/machine/start"
	netcmds "github.com/platinasystems/go/commands/net"
	vnetcmd "github.com/platinasystems/go/commands/net/vnet"
	"github.com/platinasystems/go/commands/redis"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/info/cmdline"
	"github.com/platinasystems/go/info/hostname"
	name "github.com/platinasystems/go/info/machine"
	"github.com/platinasystems/go/info/netlink"
	"github.com/platinasystems/go/info/uptime"
	"github.com/platinasystems/go/info/version"
	vnetinfo "github.com/platinasystems/go/info/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"
)

func main() {
	command.Plot(builtin.New()...)
	command.Plot(core.New()...)
	command.Plot(dlv.New()...)
	command.Plot(fs.New()...)
	command.Plot(kernel.New()...)
	command.Plot(machine.New()...)
	command.Plot(netcmds.New()...)
	command.Plot(redis.New()...)
	command.Plot(vnetcmd.New())
	command.Sort()
	start.Hook = func() error {
		if len(os.Getenv("REDISD")) == 0 {
			return nil
		}
		return os.Setenv("REDISD", "lo eth0")
	}
	machined.Hook = func() error {
		machined.Plot(
			cmdline.New(),
			hostname.New(),
			name.New("platina-mk1"),
			netlink.New(),
			uptime.New(),
			version.New(),
			vnetinfo.New(vnetinfo.Config{
				UnixInterfacesOnly: true,
				PublishAllCounters: false,
				GdbWait:            gdbwait,
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
