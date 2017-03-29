// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/platinasystems/go/internal/fdt"
	"github.com/platinasystems/go/internal/fdtgpio"
	"github.com/platinasystems/go/internal/gpio"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/led"
)

func Init() (err error) {
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
