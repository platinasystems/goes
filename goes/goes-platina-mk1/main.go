// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"os/exec"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/builtin"
	"github.com/platinasystems/go/goes/builtin/license"
	"github.com/platinasystems/go/goes/builtin/patents"
	"github.com/platinasystems/go/goes/core"
	"github.com/platinasystems/go/goes/fs"
	"github.com/platinasystems/go/goes/kernel"
	"github.com/platinasystems/go/goes/machine"
	"github.com/platinasystems/go/goes/machine/start"
	"github.com/platinasystems/go/goes/machine/stop"
	"github.com/platinasystems/go/goes/net"
	"github.com/platinasystems/go/goes/net/nld"
	"github.com/platinasystems/go/goes/net/vnet"
	"github.com/platinasystems/go/goes/net/vnetd"
	"github.com/platinasystems/go/goes/redis"
	"github.com/platinasystems/go/goes/test"
	goredis "github.com/platinasystems/go/redis"
	"github.com/platinasystems/go/sockfile"
	govnet "github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1"
	fe1copyright "github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/copyright"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"
)

const UsrShareGoes = "/usr/share/goes"

func main() {
	const fe1path = "github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1"
	license.Others = []license.Other{{fe1path, fe1copyright.License}}
	patents.Others = []patents.Other{{fe1path, fe1copyright.Patents}}
	command.Plot(builtin.New()...)
	command.Plot(core.New()...)
	command.Plot(fs.New()...)
	command.Plot(kernel.New()...)
	command.Plot(machine.New()...)
	command.Plot(net.New()...)
	command.Plot(redis.New()...)
	// command.Plot(test.New()...)
	_ = test.New
	command.Plot(vnet.New(), vnetd.New())
	command.Sort()
	start.Machine = "platina-mk1"
	start.RedisDevs = []string{"lo", "eth0"}
	start.ConfHook = wait4vnet
	stop.Hook = stopHook
	nld.Prefixes = []string{"lo.", "eth0."}
	vnetd.UnixInterfacesOnly = true
	vnetd.PublishAllCounters = false
	vnetd.GdbWait = gdbwait
	vnetd.Hook = vnetHook
	goes.Main()
}

func stopHook() error {
	for port := 0; port < 32; port++ {
		for subport := 0; subport < 4; subport++ {
			exec.Command("/bin/ip", "link", "delete",
				fmt.Sprintf("eth-%d-%d", port, subport),
			).Run()
		}
	}
	for port := 0; port < 2; port++ {
		exec.Command("/bin/ip", "link", "delete",
			fmt.Sprintf("ixge2-0-%d", port),
		).Run()
	}
	for port := 0; port < 2; port++ {
		exec.Command("/bin/ip", "link", "delete",
			fmt.Sprintf("meth-%d", port),
		).Run()
	}
	return nil
}

func vnetHook(i *vnetd.Info, v *govnet.Vnet) error {
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

func wait4vnet() error {
	conn, err := sockfile.Dial("redisd")
	if err != nil {
		return err
	}
	defer conn.Close()
	psc := redigo.PubSubConn{redigo.NewConn(conn, 0, 500*time.Millisecond)}
	if err = psc.Subscribe(goredis.Machine); err != nil {
		return err
	}
	for {
		switch t := psc.Receive().(type) {
		case redigo.Message:
			if string(t.Data) == "vnet.ready: true" {
				return nil
			}
		case error:
			return t
		}
	}
	return nil
}
