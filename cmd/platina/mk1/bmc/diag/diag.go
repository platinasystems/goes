// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package diag

import (
	"fmt"
	"sync"

	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/flags"
)

var debug, x86, writeField, delField, writeSN bool
var argF []string
var flagF *flags.Flags

type Command struct {
	Gpio func()
	gpio sync.Once
}

type Diag func() error

func (*Command) String() string { return "diag" }

func (*Command) Usage() string {
	return `
diag [-debug] | prom [-w | -delete | -x86] \
	[TYPE | "crc" | "length" | "onie" | "copy" ] [VALUE]`
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "run diagnostics",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Runs diagnostic tests to validate BMC functionality and interfaces

	EEPROM writing utility with diag prom

OPTIONS
	-x86	executes command on host EEPROM

	-w 	write flag with the following arguments
	crc 	recalculates and updates crc field
	onie 	erases contents and adds an ONIE header with crc field
	length 	debug tool to write VALUE into length field
	copy	copies host eeprom contents, updates PPN field,
		recalculates crc (vice versa with -x86)
	TYPE VALUE
		debug tool to write ONIE field of TYPE with VALUE
	-delete	delete flag with the following arguments
	TYPE	delete the first ONIE field found with TYPE

EXAMPLES
	diag prom		dumps bmc eeprom
	diag prom -x86		dumps host eeprom
	diag prom -w copy	copies host to bmc eeprom
	diag prom -x86 -w crc	updates host eeprom crc field`,
	}
}

func (c *Command) Main(args ...string) error {
	var diag string
	flagF, args = flags.New(args, "-debug", "-x86", "-w", "-delete")
	debug = flagF.ByName["-debug"]
	x86 = flagF.ByName["-x86"]
	writeField = flagF.ByName["-w"]
	delField = flagF.ByName["-delete"]
	writeSN = flagF.ByName["-wsn"]
	argF = args
	//if n := len(args); n > 1 {
	//	return fmt.Errorf("%v: unexpected", args[1:])
	//}
	if n := len(args); n != 0 {
		diag = args[0]
	}
	c.gpio.Do(c.Gpio)
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
			diagMem,
			/*
				diagNetwork,
				diagUSB,
				diagMFGProm,
			*/
		},
		"i2c":           []Diag{diagI2c},
		"uart":          []Diag{},
		"host":          []Diag{diagHost},
		"network":       []Diag{diagNetwork},
		"power":         []Diag{diagPower},
		"powerlog":      []Diag{diagLoggedFaults},
		"mem":           []Diag{diagMem},
		"usb":           []Diag{diagUSB},
		"psu":           []Diag{diagPSU},
		"fans":          []Diag{diagFans},
		"eeprom":        []Diag{diagMFGProm},
		"led":           []Diag{diagLED},
		"switchconsole": []Diag{diagSwitchConsole},
		"prom":          []Diag{diagProm},
		"powercycle":    []Diag{diagPowerCycle},
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
	var result bool
	var r string


	/* diagTest: DRAM
	tbd: run memory diag
	*/



	/* diagTest: uSD
	tbd: write/read/verify a file
	*/
	fmt.Printf("\n%15s|%25s|%10s|%10s|%10s|%10s|%6s|%35s\n", "function", "parameter", "units", "value", "min", "max", "result", "description")
	fmt.Printf("---------------|-------------------------|----------|----------|----------|----------|------|-----------------------------------\n")

	result, _ = diagDetectMMC()
	r = CheckPassB(result, true)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "uSD", "/dev/mmcblk0", "-", result, active_high_on_min, active_high_on_max, r, "check uSD present")

	// test when detected only
	if r == "pass" {
		result, _ = diagTestMMC()
	        r = CheckPassB(result, true)
		fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "uSD", "/dev/mmcblk0p1", "-", result, active_high_on_min, active_high_on_max, r, "write and read uSD")
	}



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
