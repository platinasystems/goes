// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/platinasystems/go/internal/eeprom"
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/optional/gpio"
	"github.com/platinasystems/go/internal/optional/i2c"
	"github.com/platinasystems/go/internal/optional/platina-mk1/toggle"
	"github.com/platinasystems/go/internal/optional/vnet"
	"github.com/platinasystems/go/internal/optional/vnetd"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/required"
	"github.com/platinasystems/go/internal/required/license"
	"github.com/platinasystems/go/internal/required/nld"
	"github.com/platinasystems/go/internal/required/patents"
	"github.com/platinasystems/go/internal/required/redisd"
	"github.com/platinasystems/go/internal/required/start"
	"github.com/platinasystems/go/internal/required/stop"
	govnet "github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/copyright"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"
)

const UsrShareGoes = "/usr/share/goes"

func main() {
	const fe1path = "github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1"
	license.Others = []license.Other{{fe1path, copyright.License}}
	patents.Others = []patents.Other{{fe1path, copyright.Patents}}
	g := make(goes.ByName)
	g.Plot(required.New()...)
	g.Plot(gpio.New(), i2c.New(), toggle.New(), vnet.New(), vnetd.New())
	redisd.Machine = "platina-mk1"
	redisd.Devs = []string{"lo", "eth0"}
	redisd.Hook = getEepromData
	start.ConfHook = func() error {
		return redis.Hwait(redis.DefaultHash, "vnet.ready", "true",
			10*time.Second)
	}
	stop.Hook = stopHook
	nld.Prefixes = []string{"lo.", "eth0."}
	vnetd.UnixInterfacesOnly = true
	vnetd.PublishAllCounters = false
	vnetd.GdbWait = gdbwait
	vnetd.Hook = vnetHook
	g.Main()
}

func stopHook() error {
	var startPort, endPort int

	if devEeprom.Fields.DeviceVersion == 0 {
		// Alpha level board
		startPort = 0
		endPort = 32
	} else {
		// Beta & Production level boards have version 1 and above
		startPort = 1
		endPort = 33
	}

	for port := startPort; port < endPort; port++ {
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

// The MK1 x86 CPU Card EEPROM is located on bus 0, addr 0x51:
var devEeprom = eeprom.Device{
	BusIndex:   0,
	BusAddress: 0x51,
}

func getEepromData(pub chan<- string) error {
	// Read and store the EEPROM Contents
	if err := devEeprom.GetInfo(); err != nil {
		return err
	}

	pub <- fmt.Sprint("eeprom.product_name: ", devEeprom.Fields.ProductName)
	pub <- fmt.Sprint("eeprom.platform_name: ", devEeprom.Fields.PlatformName)
	pub <- fmt.Sprint("eeprom.manufacturer: ", devEeprom.Fields.Manufacturer)
	pub <- fmt.Sprint("eeprom.vendor: ", devEeprom.Fields.Vendor)
	pub <- fmt.Sprint("eeprom.part_number: ", devEeprom.Fields.PartNumber)
	pub <- fmt.Sprint("eeprom.serial_number: ", devEeprom.Fields.SerialNumber)
	pub <- fmt.Sprint("eeprom.devEepromice_version: ", devEeprom.Fields.DeviceVersion)
	pub <- fmt.Sprint("eeprom.manufacture_date: ", devEeprom.Fields.ManufactureDate)
	pub <- fmt.Sprint("eeprom.country_code: ", devEeprom.Fields.CountryCode)
	pub <- fmt.Sprint("eeprom.diag_version: ", devEeprom.Fields.DiagVersion)
	pub <- fmt.Sprint("eeprom.service_tag: ", devEeprom.Fields.ServiceTag)
	pub <- fmt.Sprint("eeprom.base_ethernet_address: ", devEeprom.Fields.BaseEthernetAddress)
	pub <- fmt.Sprint("eeprom.number_of_ethernet_addrs: ", devEeprom.Fields.NEthernetAddress)
	return nil
}
