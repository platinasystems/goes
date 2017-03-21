// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/platinasystems/go/internal/environ/fantray"
	"github.com/platinasystems/go/internal/environ/fsp"
	"github.com/platinasystems/go/internal/environ/nuvoton"
	"github.com/platinasystems/go/internal/environ/nxp"
	"github.com/platinasystems/go/internal/environ/ti"
	"github.com/platinasystems/go/internal/fdt"
	"github.com/platinasystems/go/internal/fdtgpio"
	"github.com/platinasystems/go/internal/gpio"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/led"
)

func Init() (err error) {
	ucd9090Init()
	w83795Init()
	fantrayInit()
	imx6Init()
	fspInit()
	ledgpioInit()
	if err = boardInit(); err != nil {
		return err
	}
	return nil
}

func boardInit() (err error) {
	gpio.File = "/boot/platina-mk1-bmc.dtb"
	gpio.Aliases = make(gpio.GpioAliasMap)
	gpio.Pins = make(gpio.PinMap)

	// Parse linux.dtb to generate gpio map for this machine
	if b, err := ioutil.ReadFile(gpio.File); err == nil {
		t := &fdt.Tree{Debug: false, IsLittleEndian: false}
		t.Parse(b)

		t.MatchNode("aliases", fdtgpio.GatherAliases)
		t.EachProperty("gpio-controller", "", fdtgpio.GatherPins)
	} else {
		return fmt.Errorf("%s: %v", gpio.File, err)
	}

	return nil
}

func ledgpioInit() {
	ledgpio.Vdev.Bus = 0
	ledgpio.Vdev.Addr = 0x0 //update after eeprom read
	ledgpio.Vdev.MuxBus = 0x0
	ledgpio.Vdev.MuxAddr = 0x76
	ledgpio.Vdev.MuxValue = 0x2
	ver, _ := readVer()
	switch ver {
	case 0xff:
		ledgpio.Vdev.Addr = 0x22
	case 0x00:
		ledgpio.Vdev.Addr = 0x22
	default:
		ledgpio.Vdev.Addr = 0x75
	}
}

func ucd9090Init() {
	ucd9090.Vdev.Bus = 0
	ucd9090.Vdev.Addr = 0x0 //update after eeprom read
	ucd9090.Vdev.MuxBus = 0
	ucd9090.Vdev.MuxAddr = 0x76
	ucd9090.Vdev.MuxValue = 0x01
	ver, _ := readVer()
	switch ver {
	case 0xff:
		ucd9090.Vdev.Addr = 0x7e
	case 0x00:
		ucd9090.Vdev.Addr = 0x7e
	default:
		ucd9090.Vdev.Addr = 0x34
	}

	ucd9090.VpageByKey = map[string]uint8{
		"vmon.5v.sb.units.V":    1,
		"vmon.3v8.bmc.units.V":  2,
		"vmon.3v3.sys.units.V":  3,
		"vmon.3v3.bmc.units.V":  4,
		"vmon.3v3.sb.units.V":   5,
		"vmon.1v0.thc.units.V":  6,
		"vmon.1v8.sys.units.V":  7,
		"vmon.1v25.sys.units.V": 8,
		"vmon.1v2.ethx.units.V": 9,
		"vmon.1v0.tha.units.V":  10,
	}
}

func w83795Init() {
	w83795.Vdev.Bus = 0
	w83795.Vdev.Addr = 0x2f
	w83795.Vdev.MuxBus = 0
	w83795.Vdev.MuxAddr = 0x76
	w83795.Vdev.MuxValue = 0x80

	w83795.VpageByKey = map[string]uint8{
		"fan_tray.1.1.speed.units.rpm": 1,
		"fan_tray.1.2.speed.units.rpm": 2,
		"fan_tray.2.1.speed.units.rpm": 3,
		"fan_tray.2.2.speed.units.rpm": 4,
		"fan_tray.3.1.speed.units.rpm": 5,
		"fan_tray.3.2.speed.units.rpm": 6,
		"fan_tray.4.1.speed.units.rpm": 7,
		"fan_tray.4.2.speed.units.rpm": 8,
		"fan_tray.speed":               1,
	}
}

func fantrayInit() {
	fantray.Vdev.Bus = 1
	fantray.Vdev.Addr = 0x20
	fantray.Vdev.MuxBus = 1
	fantray.Vdev.MuxAddr = 0x72
	fantray.Vdev.MuxValue = 0x04

	fantray.VpageByKey = map[string]uint8{
		"fan_tray.1.status": 1,
		"fan_tray.2.status": 2,
		"fan_tray.3.status": 3,
		"fan_tray.4.status": 4,
	}
}

func imx6Init() {
	imx6.VpageByKey = map[string]uint8{
		"bmc.temperature.units.C": 1,
	}
}

func fspInit() {
	fsp.Vdev[0].Slot = 2
	fsp.Vdev[0].Bus = 1
	fsp.Vdev[0].Addr = 0x58
	fsp.Vdev[0].MuxBus = 1
	fsp.Vdev[0].MuxAddr = 0x72
	fsp.Vdev[0].MuxValue = 0x01
	fsp.Vdev[0].GpioPwrok = "PSU0_PWROK"
	fsp.Vdev[0].GpioPrsntL = "PSU0_PRSNT_L"
	fsp.Vdev[0].GpioPwronL = "PSU0_PWRON_L"
	fsp.Vdev[0].GpioIntL = "PSU0_INT_L"

	fsp.Vdev[1].Slot = 1
	fsp.Vdev[1].Bus = 1
	fsp.Vdev[1].Addr = 0x58
	fsp.Vdev[1].MuxBus = 1
	fsp.Vdev[1].MuxAddr = 0x72
	fsp.Vdev[1].MuxValue = 0x02
	fsp.Vdev[1].GpioPwrok = "PSU1_PWROK"
	fsp.Vdev[1].GpioPrsntL = "PSU1_PRSNT_L"
	fsp.Vdev[1].GpioPwronL = "PSU1_PWRON_L"
	fsp.Vdev[1].GpioIntL = "PSU1_INT_L"

	fsp.VpageByKey = map[string]uint8{
		"psu1.status":              1,
		"psu1.admin.state":         1,
		"psu1.mfg_id":              1,
		"psu1.mfg_model":           1,
		"psu1.v_in.units.V":        1,
		"psu1.v_out.units.V":       1,
		"psu1.p_out.units.W":       1,
		"psu1.p_in.units.W":        1,
		"psu1.temperature.units.C": 1,
		"psu2.status":              0,
		"psu2.admin.state":         0,
		"psu2.mfg_id":              0,
		"psu2.mfg_model":           0,
		"psu2.v_in.units.V":        0,
		"psu2.v_out.units.V":       0,
		"psu2.p_out.units.W":       0,
		"psu2.p_in.units.W":        0,
		"psu2.temperature.units.C": 0,
	}
}

func readVer() (v int, err error) {
	f, err := os.Open("/tmp/ver")
	if err != nil {
		return 0, err
	}
	b1 := make([]byte, 5)
	_, err = f.Read(b1)
	if err != nil {
		return 0, err
	}
	f.Close()
	return int(b1[0]), nil
}
