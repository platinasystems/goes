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
)

func Init() (err error) {
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
