// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/commands/builtin"
	"github.com/platinasystems/go/commands/core"
	"github.com/platinasystems/go/commands/dlv"
	"github.com/platinasystems/go/commands/fs"
	"github.com/platinasystems/go/commands/kernel"
	"github.com/platinasystems/go/commands/machine"
	"github.com/platinasystems/go/commands/machine/install"
	"github.com/platinasystems/go/commands/machine/machined"
	"github.com/platinasystems/go/commands/machine/start"
	netcmds "github.com/platinasystems/go/commands/net"
	vnetcmd "github.com/platinasystems/go/commands/net/vnet"
	"github.com/platinasystems/go/commands/redis"
	"github.com/platinasystems/go/firmware/fe1a"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/info/cmdline"
	"github.com/platinasystems/go/info/hostname"
	name "github.com/platinasystems/go/info/machine"
	"github.com/platinasystems/go/info/netlink"
	"github.com/platinasystems/go/info/uptime"
	"github.com/platinasystems/go/info/version"
	vnetinfo "github.com/platinasystems/go/info/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"
)

const UsrShareGoes = "/usr/share/goes"

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
	start.Hook = startHook
	machined.Hook = machinedHook
	install.Hook = installHook
	goes.Main()
}

func startHook() error {
	if len(os.Getenv("REDISD")) == 0 {
		return nil
	}
	return os.Setenv("REDISD", "lo eth0")
}

func machinedHook() error {
	err := fe1a.Load()
	if err != nil {
		return err
	}
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

func installHook() error {
	_, err := os.Stat(UsrShareGoes)
	if os.IsNotExist(err) {
		err = os.Mkdir(UsrShareGoes, os.FileMode(0755))
		if err != nil {
			return err
		}
	}
	for _, fn := range []string{"tsce.ucode", "tscf.ucode"} {
		var src, dst *os.File
		for _, dir := range []string{
			".",
			"firmware/fe1a",
			"src/github.com/platinasystems/go/firmware/fe1a",
		} {
			src, err = os.Open(filepath.Join(dir, fn))
			if err == nil {
				break
			}
		}
		if err != nil {
			return fmt.Errorf("%s: not found")
		}
		dst, err = os.Create(filepath.Join(UsrShareGoes, fn))
		if err != nil {
			src.Close()
			return err
		}
		_, err = io.Copy(dst, src)
		src.Close()
		dst.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func vnetHook(i *vnetinfo.Info) error {
	v := i.V()

	// Base packages.
	ethernet.Init(v)
	ip4.Init(v)
	ip6.Init(v)
	pg.Init(v)   // vnet packet generator
	unix.Init(v) // tuntap/netlink

	// Device drivers: FE1 switch + Intel 10G ethernet for punt path.
	ixge.Init(v)
	fe1.Init(v)

	plat := &platform{i: i}
	v.AddPackage("platform", plat)
	plat.DependsOn("pci-discovery")

	return nil
}
