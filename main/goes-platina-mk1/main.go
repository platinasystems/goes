// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/optional/eeprom"
	"github.com/platinasystems/go/internal/optional/eeprom/platina_eeprom"
	"github.com/platinasystems/go/internal/optional/gpio"
	"github.com/platinasystems/go/internal/optional/i2c"
	"github.com/platinasystems/go/internal/optional/i2cd"
	"github.com/platinasystems/go/internal/optional/platina-mk1/toggle"
	"github.com/platinasystems/go/internal/optional/vnet"
	"github.com/platinasystems/go/internal/optional/vnetd"
	"github.com/platinasystems/go/internal/prog"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/required"
	"github.com/platinasystems/go/internal/required/license"
	"github.com/platinasystems/go/internal/required/patents"
	"github.com/platinasystems/go/internal/required/redisd"
	"github.com/platinasystems/go/internal/required/start"
	"github.com/platinasystems/go/internal/required/stop"
	govnet "github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/copyright"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/firmware"
	"github.com/platinasystems/go/vnet/devices/optics/qsfp"
	"github.com/platinasystems/go/vnet/devices/optics/qsfpio"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"
)

const UsrShareGoes = "/usr/share/goes"

func main() {
	var h platform
	const fe1path = "github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1"
	license.Others = []license.Other{{fe1path, copyright.License}}
	patents.Others = []patents.Other{{fe1path, copyright.Patents}}
	g := make(goes.ByName)
	g.Plot(required.New()...)
	g.Plot(
		eeprom.New(),
		gpio.New(),
		i2c.New(),
		i2cd.New(),
		qsfp.New(),
		qsfpio.New(),
		toggle.New(),
		vnet.New(),
		vnetd.New(),
	)
	redisd.Machine = "platina-mk1"
	redisd.Devs = []string{"lo", "eth0"}
	h.QsfpioInit()
	h.QsfpInit()
	redisd.Hook = platina_eeprom.RedisdHook
	start.ConfHook = func() error {
		return redis.Hwait(redis.DefaultHash, "vnet.ready", "true",
			10*time.Second)
	}
	stop.Hook = stopHook
	vnetd.UnixInterfacesOnly = true
	vnetd.GdbWait = gdbwait
	vnetd.Hook = vnetHook
	platina_eeprom.Config(
		platina_eeprom.BusIndex(0),
		platina_eeprom.BusAddress(0x51),
		platina_eeprom.BusDelay(10*time.Millisecond),
		platina_eeprom.MinMacs(134),
		platina_eeprom.OUI([3]byte{0x02, 0x46, 0x8a}),
	)
	if err := g.Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func stopHook() error {
	// Alpha: 0:32
	// Beta: 1:33
	// So, cover all with 0..33
	for port := 0; port < 33; port++ {
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
	err := firmware.Extract(prog.Name())
	if err != nil {
		return err
	}

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
