// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This is an example Baseboard Management Controller.
package main

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/internal/eeprom"
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/optional/i2c"
	"github.com/platinasystems/go/internal/optional/platina-mk1/toggle"
	"github.com/platinasystems/go/internal/optional/telnetd"
	"github.com/platinasystems/go/internal/optional/watchdog"
	//	"github.com/platinasystems/go/internal/prog"
	//	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/gpio"
	optgpio "github.com/platinasystems/go/internal/optional/gpio"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/diag"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ"
	"github.com/platinasystems/go/internal/required"
	"github.com/platinasystems/go/internal/required/nld"
	"github.com/platinasystems/go/internal/required/redisd"
	"github.com/platinasystems/go/internal/required/start"
	"github.com/platinasystems/go/internal/required/stop"
)

const UsrShareGoes = "/usr/share/goes"

func main() {
	gpio.File = "/boot/platina-mk1-bmc.dtb"
	g := make(goes.ByName)
	g.Plot(required.New()...)
	g.Plot(diag.New(), optgpio.New(), i2c.New(), telnetd.New(), toggle.New(), watchdog.New())
	g.Plot(environ.New()...)
	redisd.Machine = "platina-mk1-bmc"
	redisd.Devs = []string{"lo", "eth0"}
	redisd.Hook = getEepromData
	start.ConfHook = func() error {
		return nil
	}
	stop.Hook = stopHook
	nld.Prefixes = []string{"lo.", "eth0."}
	if err := g.Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func stopHook() error {
	return nil
}

// The MK1 x86 CPU Card EEPROM is located on bus 0, addr 0x51:
var devEeprom = eeprom.Device{
	BusIndex:   0,
	BusAddress: 0x51,
}

func getEepromData(pub chan<- string) error {
	// Read and store the EEPROM Contents
	/*
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
	*/
	return nil
}
