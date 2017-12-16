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
)

func gpioInit() {
	gpio.File = "/boot/platina-mk2-mc1-bmc.dtb"
	gpio.Aliases = make(gpio.GpioAliasMap)
	gpio.Pins = make(gpio.PinMap)

	// Parse linux.dtb to generate gpio map for this machine
	b, err := ioutil.ReadFile(gpio.File)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", gpio.File, err)
		return
	}
	t := &fdt.Tree{Debug: false, IsLittleEndian: false}
	t.Parse(b)

	t.MatchNode("aliases", fdtgpio.GatherAliases)
	t.EachProperty("gpio-controller", "", fdtgpio.GatherPins)
}
