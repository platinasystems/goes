// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package gpio provides cli command to query/configure GPIO pins.
package gpio

import (
	"fmt"
	"sort"

	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/gpio"
)

type Command struct{}

func (Command) String() string { return "gpio" }

func (Command) Usage() string { return "gpio PIN_NAME [VALUE]" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "manipulate GPIO pins",
	}
}

func (c Command) Main(args ...string) error {
	switch len(args) {
	case 0: // No args?  Report all pin values.
		names := make([]string, 0, gpio.NumPins())
		for name, _ := range gpio.AllPins() {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			pin, _ := gpio.FindPin(name)
			v, err := pin.Value()
			if err != nil {
				fmt.Printf("%s: %v\n", name, err)
			}
			fmt.Printf("%s: %v\n", name, v)
		}
	case 1:
		if args[0] == "default" {
			// Set pin directions.
			for name, pin := range gpio.AllPins() {
				err := pin.SetDefault()
				if err != nil {
					fmt.Printf("%s: %v\n", name, err)
				}
			}
		} else {
			pin, found := gpio.FindPin(args[0])
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
		pin, found := gpio.FindPin(args[0])
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
