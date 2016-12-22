// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package gpio provides cli command to query/configure GPIO pins.
package gpio

import (
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/platinasystems/go/goes/internal/fdt"
	"github.com/platinasystems/go/goes/internal/fdtgpio"
	"github.com/platinasystems/go/gpio"
)

const Name = "gpio"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " PIN_NAME [VALUE]" }

func (cmd) Main(args ...string) error {
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

	switch len(args) {
	case 0: // No args?  Report all pin values.
		names := make([]string, 0, len(gpio.Pins))
		for name, _ := range gpio.Pins {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			pin := gpio.Pins[name]
			v, err := pin.Value()
			if err != nil {
				fmt.Printf("%s: %v\n", name, err)
			}
			fmt.Printf("%s: %v\n", name, v)
		}
	case 1:
		if args[0] == "default" {
			// Set pin directions.
			for name, pin := range gpio.Pins {
				err := pin.SetDirection()
				if err != nil {
					fmt.Printf("%s: %v\n", name, err)
				}
			}
		} else {
			pin, found := gpio.Pins[args[0]]
			if !found {
				return fmt.Errorf("%s: not found", args[0])
			}
			v, err := pin.Value()
			if err != nil {
				fmt.Printf("%s: %v\n", args[0], err)
			}
			fmt.Printf("%s: %v\n", args[0], v)
		}
	case 2:
		pin, found := gpio.Pins[args[0]]
		if !found {
			return fmt.Errorf("%s: not found", args[0])
		}
		switch args[1] {
		case "true", "1":
			return pin.SetValue(true)
		case "false", "0":
			return pin.SetValue(false)
		default:
			return fmt.Errorf("%s: invalid, must be true|false",
				args[1])
		}
	default:
		return fmt.Errorf("%v: unexpected", args[2:])
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
