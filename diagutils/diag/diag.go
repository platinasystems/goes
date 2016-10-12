// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package diag

import (
	"fmt"
	"io/ioutil"

	"github.com/platinasystems/go/fdt"
	"github.com/platinasystems/go/fdtgpio"
	"github.com/platinasystems/go/gpio"
)

type diag struct{}

func New() diag { return diag{} }

func (diag) String() string { return "diag" }
func (diag) Usage() string  { return "diag" }

func diagMem() {
	/* diagTest: DRAM
	tbd: run memory diag
	*/

	/* diagTest: uSD
	tbd: write/read/verify a file
	*/

	/* diagTest: QSPI
	tbd: likely n/a QSPI tested via SW upgrade path, need to validate dual boot if supported
	*/
}

func diagUSB() {
	/* diagTest: USB
	tbd: write/read/verify a file
	*/
	//select BMC USB on front panel
	//pv := gpio.PinValue{Name: "USB_MUX_SEL"}
	//pv.PinNum.SetValue(true)

}

func diagMFGProm() {
	/* diagTest: MFG EEPROM
	   tbd: dump eeprom fields
	   tbd: dump platina portion only
	   tbd: dump entire eeprom
	   tbd: write each field individually
	   tbd: read each field individually
	*/
}

func diagLED() {
	/* diagTest: Front panel LEDs
	   tbd: toggle LEDs in a sequence for operator to check
	   tbd: use PSU_PWROK signal to validate PSU leds
	*/
}

var debug bool
var diagGpios gpio.PinMap

func (diag) Main(args ...string) error {

	debug = false

	if len(args) > 2 || len(args) == 0 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	//create GPIO map
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
	//diagGpios = gpio.Pins

	if len(args) > 1 {
		if args[1] == "debug" {
			debug = true
		}
	}

	if args[0] == "all" {
		diagI2c()
		diagPower()
		diagFans()
		diagPSU()
		diagHost()
		diagNetwork()
		// diagUSB()
		//diagMem()
		//diagMFGProm()

	} else if args[0] == "i2c" {
		diagI2c()
	} else if args[0] == "uart" {

	} else if args[0] == "host" {
		diagHost()
	} else if args[0] == "network" {
		diagNetwork()
	} else if args[0] == "power" {
		diagPower()
	} else if args[0] == "mem" {
		diagMem()
	} else if args[0] == "usb" {
		diagUSB()
	} else if args[0] == "psu" {
		diagPSU()
	} else if args[0] == "fans" {
		diagFans()
	} else if args[0] == "mfg_eeprom" {
		diagMFGProm()
	} else if args[0] == "led" {
		diagLED()
	} else {
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	//power monitoring
	//UARt
	//BMC and Host Interaction
	//Ethernet
	//Memories
	//USB
	//PSU
	//Fans
	//MFG EEPROM

	//pinstate := 0

	/*
		// Print Header
		fmt.Printf("%-10s:%-20s:%-10s:%-10s:%-10s:%-10s:%-5s\n","function","parameter","units","value","min","max","result")

		// LTC4215 Diagnostics
		// read output voltage
		fn := "/sys/bus/i2c/devices/4-0048/hwmon/hwmon0/in2_input"
		f, err := goes.OpenURL(fn)
		if err != nil {
			diag.Panic(err)
		}
		buf := make([]byte,10)
		n, err := io.ReadAtLeast(f, buf,1)
		if err != nil {
			diag.Panic(err)
		}
		i,err := strconv.Atoi(string(buf[:(n-1)]))
		v := float64(i)/1000
		f.Close()
		min := 11.000
		max := 13.000
		r := "fail"
		if v >= min && v <= max {
			r = "pass"
		}

		fmt.Printf("%-10s:%-20s:%-10s:%-10f:%-10f:%-10f:%-5s\n","LTC4215","output_voltage","V",v,min,max,r)
	*/
	return nil
}

func (diag) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "run diagnostics",
	}
}

func (diag) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	diag - run diagnostics

SYNOPSIS
	diag

DESCRIPTION
	Runs diagnostic tests to validate BMC functionality and interfaces

EXAMPLES
	diag`,
	}
}
