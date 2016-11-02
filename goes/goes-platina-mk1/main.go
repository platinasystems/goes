// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package main

import (
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
	"github.com/platinasystems/go/machined/info"
	"github.com/platinasystems/go/machined/info/cmdline"
	"github.com/platinasystems/go/machined/info/hostname"
	"github.com/platinasystems/go/machined/info/netlink"
	"github.com/platinasystems/go/machined/info/uptime"
	"github.com/platinasystems/go/machined/info/version"
	"github.com/platinasystems/go/machined/info/vnetinfo"
	"github.com/platinasystems/go/netutils"
	"github.com/platinasystems/go/redisutils"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"

	"os"
	"sync"
)

type platform struct {
	vnet.Package
	*bcm.Platform
	i *Info
}

type Info struct {
	mutex sync.Mutex
	name  string
	v     *vnet.Vnet
	vi    *vnetinfo.Info
}

func main() {
	command.Plot(builtinutils.New()...)
	command.Plot(coreutils.New()...)
	command.Plot(dlv.New()...)
	command.Plot(diagutils.New()...)
	command.Plot(fsutils.New()...)
	command.Plot(initutils.New()...)
	command.Plot(kutils.New()...)
	command.Plot(machined.New())
	command.Plot(vnetinfo.NewCmd())
	command.Plot(netutils.New()...)
	command.Plot(redisutils.New()...)
	command.Sort()
	start.Hook = func() error {
		os.Setenv("REDISD", "lo eth0")
		return nil
	}
	machined.Hook = hook
	goes.Main()
}

func hook() error {
	v := &vnet.Vnet{}
	i := &Info{
		name: "mk1",
	}
	i.v = v
	i.vi = vnetinfo.New(v, vnetinfo.Config{
		UnixInterfacesOnly: false,
		PublishAllCounters: true,
		GdbWait:            false,
	})
	machined.Plot(
		cmdline.New(),
		hostname.New(),
		netlink.New(),
		uptime.New(),
		version.New(),
		i.vi,
		i)
	machined.Info["netlink"].Prefixes("lo.", "eth0.")
	return nil
}

func (p *Info) String() string { return p.name }

func (i *Info) Main(...string) error {
	// Public machine name.
	info.Publish("machine", "platina-mk1")

	// Publish units.
	for _, entry := range []struct{ name, unit string }{
		{"current", "milliamperes"},
		{"fan", "% max speed"},
		{"potential", "volts"},
		{"temperature", "°C"},
	} {
		info.Publish("unit."+entry.name, entry.unit)
	}

	i.configureVnet()
	return i.vi.Start()
}

func (i *Info) configureVnet() {
	v := i.v

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
}

func (*Info) Close() error                         { return nil }
func (*Info) Set(key, value string) error          { return info.CantSet(key) }
func (*Info) Del(key string) error                 { return info.CantDel(key) }
func (*Info) Prefixes(prefixes ...string) []string { return []string{} }
