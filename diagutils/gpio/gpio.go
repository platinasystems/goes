// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package diagutils/gpio provides a command to query and configure GPIO pins.
package gpio

import (
	"github.com/platinasystems/go/diagutils/internal"
	"github.com/platinasystems/go/fdt"
	"github.com/platinasystems/go/fdtgpio"
	"github.com/platinasystems/go/gpio"

	"fmt"
	"io/ioutil"
	"sort"
)

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return "gpio" }
func (cmd) Usage() string  { return "gpio PIN_NAME [VALUE]" }

func (cmd) Main(args ...string) error {
	if len(args) > 2 {
		return fmt.Errorf("%v: unexpected", args[2:])
	}

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

	if len(args) == 1 && args[0] == "default" {
		// Set pin directions.
		for _, p := range gpio.Pins {
			err := p.SetDirection()
			if err != nil {
				// Don't panic just report error and continue.
				fmt.Println(err)
			}
		}
		return nil
	}

	// No args?  Report all pin values.
	if len(args) == 0 {

		pvs := internal.PinValues{}
		for n, p := range gpio.Pins {
			v, err := p.Value()
			if err != nil {
				// Don't panic just report error and continue.
				fmt.Println(err)
			}
			pvs = append(pvs, internal.PinValue{
				Name:   n,
				Value:  v,
				PinNum: p,
			})
		}
		sort.Sort(pvs)
		for i := range pvs {
			fmt.Println(&pvs[i])
		}
		return nil
	}

	if len(args) > 0 {
		pv := internal.PinValue{Name: args[0]}
		ok := false
		if pv.PinNum, ok = gpio.Pins[pv.Name]; !ok {
			return fmt.Errorf("no such pin: %s", pv.Name)
		}

		if len(args) == 2 {
			switch args[1] {
			case "true", "1":
				ok = true
			case "false", "0":
				ok = false
			default:
				return fmt.Errorf("expected true|false or 0|1, got %s",
					args[1])
			}
			pv.PinNum.SetValue(ok)
		}

		// Single arg? Read specified pin value.
		if v, err := pv.PinNum.Value(); err != nil {
			fmt.Println(err)
		} else {
			pv.Value = v
		}
		fmt.Println(&pv)
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "manipulate GPIO pins",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	gpio - Manipulate GPIO pins

SYNOPSIS
	gpio

DESCRIPTION
	Manipulate GPIO pins`,
	}
}
