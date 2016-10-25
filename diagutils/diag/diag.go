// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package diag

import (
	"fmt"
	"io/ioutil"

	"github.com/platinasystems/go/fdt"
	"github.com/platinasystems/go/fdtgpio"
	"github.com/platinasystems/go/flags"
	"github.com/platinasystems/go/gpio"
)

const Name = "diag"

var debug bool

type cmd struct{}
type Diag func() error

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [-debug]" }

func (cmd) Main(args ...string) error {
	var diag string
	flag, args := flags.New(args, "-debug")
	debug = flag["-debug"]
	if n := len(args); n > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	} else if n == 1 {
		diag = args[0]
	}

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

	diags, found := map[string][]Diag{
		"": []Diag{
			diagI2c,
			diagPower,
			diagFans,
			diagPSU,
			diagHost,
		},
		"all": []Diag{
			diagI2c,
			diagPower,
			diagFans,
			diagPSU,
			diagHost,
			/*
				diagNetwork,
				diagUSB,
				diagMem,
				diagMFGProm,
			*/
		},
		"i2c":     []Diag{diagI2c},
		"uart":    []Diag{},
		"host":    []Diag{diagHost},
		"network": []Diag{diagNetwork},
		"power":   []Diag{diagPower},
		"mem":     []Diag{diagMem},
		"usb":     []Diag{diagUSB},
		"psu":     []Diag{diagPSU},
		"fans":    []Diag{diagFans},
		"eeprom":  []Diag{diagMFGProm},
		"led":     []Diag{diagLED},
	}[diag]
	if !found {
		return fmt.Errorf("%s: unknown", diag)
	}
	if len(diags) == 0 {
		return fmt.Errorf("%s: unavailable", diag)
	}
	for _, f := range diags {
		if err := f(); err != nil {
			return err
		}
	}
	return nil

}

func diagMem() error {
	/* diagTest: DRAM
	tbd: run memory diag
	*/

	/* diagTest: uSD
	tbd: write/read/verify a file
	*/

	/* diagTest: QSPI
	tbd: likely n/a QSPI tested via SW upgrade path, need to validate dual boot if supported
	*/
	return nil
}

func diagUSB() error {
	/* diagTest: USB
	tbd: write/read/verify a file
	*/
	//select BMC USB on front panel
	//pv := gpio.PinValue{Name: "USB_MUX_SEL"}
	//pv.PinNum.SetValue(true)

	return nil
}

func diagMFGProm() error {
	/* diagTest: MFG EEPROM
	   tbd: dump eeprom fields
	   tbd: dump platina portion only
	   tbd: dump entire eeprom
	   tbd: write each field individually
	   tbd: read each field individually
	*/
	return nil
}

func diagLED() error {
	/* diagTest: Front panel LEDs
	   tbd: toggle LEDs in a sequence for operator to check
	   tbd: use PSU_PWROK signal to validate PSU leds
	*/
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "run diagnostics",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	diag - run diagnostics

SYNOPSIS
	diag [-debug] [DIAG | "all"]

DESCRIPTION
	Runs diagnostic tests to validate BMC functionality and interfaces

EXAMPLES
	diag`,
	}
}
