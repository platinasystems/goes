// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This is an example Baseboard Management Controller.
package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/platinasystems/go/internal/eeprom"
	"github.com/platinasystems/go/internal/environ/fantray"
	"github.com/platinasystems/go/internal/environ/fsp"
	"github.com/platinasystems/go/internal/environ/nuvoton"
	"github.com/platinasystems/go/internal/environ/nxp"
	"github.com/platinasystems/go/internal/environ/ti"
	"github.com/platinasystems/go/internal/fdt"
	"github.com/platinasystems/go/internal/fdtgpio"
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/gpio"
	optgpio "github.com/platinasystems/go/internal/optional/gpio"
	"github.com/platinasystems/go/internal/optional/i2c"
	"github.com/platinasystems/go/internal/optional/i2cd"
	"github.com/platinasystems/go/internal/optional/platina-mk1/toggle"
	"github.com/platinasystems/go/internal/optional/telnetd"
	"github.com/platinasystems/go/internal/optional/watchdog"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/diag"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/led"
	"github.com/platinasystems/go/internal/required"
	"github.com/platinasystems/go/internal/required/redisd"
	"github.com/platinasystems/go/internal/required/start"
	"github.com/platinasystems/go/internal/required/stop"
)

const UsrShareGoes = "/usr/share/goes"

func main() {
	gpio.File = "/boot/platina-mk1-bmc.dtb"
	var h platform
	g := make(goes.ByName)
	g.Plot(required.New()...)
	g.Plot(
		diag.New(),
		optgpio.New(),
		i2c.New(),
		telnetd.New(),
		toggle.New(),
		watchdog.New(),
		i2cd.New(),
		fantray.New(),
		fsp.New(),
		imx6.New(),
		w83795.New(),
		ucd9090.New(),
		ledgpio.New(),
	)
	redisd.Machine = "platina-mk1-bmc"
	redisd.Devs = []string{"lo", "eth0"}
	/* FIXME
	redisd.Hook = platina_eeprom.RedisdHook
	platina_eeprom.Config(
		platina_eeprom.BusIndex(0),
		platina_eeprom.BusAddress(0x51),
		platina_eeprom.BusDelay(10*time.Millisecond),
		platina_eeprom.MinMacs(2),
		platina_eeprom.OUI([3]byte{0x02, 0x46, 0x8a}),
	)
	*/
	if err := h.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	start.ConfGpioHook = func() error {
		gpio.Aliases = make(gpio.GpioAliasMap)
		gpio.Pins = make(gpio.PinMap)
		if b, err := ioutil.ReadFile(gpio.File); err == nil {
			t := &fdt.Tree{Debug: false, IsLittleEndian: false}
			t.Parse(b)

			t.MatchNode("aliases", fdtgpio.GatherAliases)
			t.EachProperty("gpio-controller", "", fdtgpio.GatherPins)
		} else {
			return fmt.Errorf("%s: %v", gpio.File, err)
		}
		for name, pin := range gpio.Pins {
			err := pin.SetDirection()
			if err != nil {
				fmt.Printf("%s: %v\n", name, err)
			}
		}
		return nil
	}
	start.ConfHook = func() error {
		return nil
	}
	stop.Hook = stopHook
	if err := g.Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
