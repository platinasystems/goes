// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"

	"github.com/platinasystems/go/internal/fdt"
	"github.com/platinasystems/go/internal/fdtgpio"
	"github.com/platinasystems/go/internal/gpio"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ/nuvoton"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ/ti"
)

type platform struct {
}

func (p *platform) Init() (err error) {
	if err = p.ucd9090Init(); err != nil {
		return err
	}
	if err = p.w83795Init(); err != nil {
		return err
	}
	if err = p.boardInit(); err != nil {
		return err
	}
	return nil
}

func (p *platform) boardInit() (err error) {
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

	// Set gpio input/output as defined in dtb
	for name, pin := range gpio.Pins {
		err := pin.SetDirection()
		if err != nil {
			fmt.Printf("%s: %v\n", name, err)
		}
	}
	return nil
}

func (p *platform) ucd9090Init() (err error) {
	ucd9090.Vdev.Bus = 0
	ucd9090.Vdev.Addr = 0x7e
	ucd9090.Vdev.MuxBus = 0
	ucd9090.Vdev.MuxAddr = 0x76
	ucd9090.Vdev.MuxValue = 0x01

	ucd9090.VpageByKey = map[string]uint8{
		"vmon.5v.sb":    1,
		"vmon.3v8.bmc":  2,
		"vmon.3v3.sys":  3,
		"vmon.3v3.bmc":  4,
		"vmon.3v3.sb":   5,
		"vmon.1v0.thc":  6,
		"vmon.1v8.sys":  7,
		"vmon.1v25.sys": 8,
		"vmon.1v2.ethx": 9,
		"vmon.1v0.tha":  10,
	}

	return nil
}

func (p *platform) w83795Init() (err error) {
	w83795.Vdev.Bus = 0
	w83795.Vdev.Addr = 0x2f
	w83795.Vdev.MuxBus = 0
	w83795.Vdev.MuxAddr = 0x76
	w83795.Vdev.MuxValue = 0x80

	w83795.VpageByKey = map[string]uint8{
		"fan_tray.1.1.rpm": 1,
		"fan_tray.1.2.rpm": 2,
		"fan_tray.2.1.rpm": 3,
		"fan_tray.2.2.rpm": 4,
		"fan_tray.3.1.rpm": 5,
		"fan_tray.3.2.rpm": 6,
		"fan_tray.4.1.rpm": 7,
		"fan_tray.4.2.rpm": 8,
	}

	return nil
}
