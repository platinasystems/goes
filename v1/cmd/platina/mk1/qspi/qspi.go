// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package qspi

import (
	"fmt"
	"strconv"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/ioport"
)

type Command struct{}

func (Command) String() string { return "qspi" }

func (Command) Usage() string {
	return "qspi [UNIT]"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "set or return selected QSPI Flash",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	The qspi command reports or sets the active QSPI device.

USAGE
	qspi [-watchdog] unit

	-watchdog	Enable the watchdog timer

	Specify the primary QSPI as 0 and the backup as 1.`,
	}
}

func (c Command) Main(args ...string) (err error) {
	flag, args := flags.New(args,
		"-watchdog",
	)

	sel := 0
	s1, err := ioport.Inb(0x602)
	if err != nil {
		return fmt.Errorf("Error reading CPLD Status-1: %s", err)
	}
	if s1&0x80 != 0 {
		sel = 1
	}

	if len(args) == 0 {
		fmt.Printf("QSPI%d is selected\n", sel)
		if !flag.ByName["-watchdog"] {
			return nil
		}
	}

	req := sel
	if len(args) > 0 {
		req, err = strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("Error parsing %s: %s", args[0], err)
		}
		args = args[1:]
		if len(args) > 0 {
			return fmt.Errorf("%v: unexpected", args)
		}
		if req < 0 || req > 1 {
			return fmt.Errorf("Invalid unit number %d", req)
		}
	}

	c1, err := ioport.Inb(0x604)
	if err != nil {
		return fmt.Errorf("Error reading CPLD Ctrl-1: %s", err)
	}
	c1 = (c1 & 0x7c) | byte(sel<<7)
	err = ioport.Outb(0x604, c1)
	if err != nil {
		return fmt.Errorf("Error clearing watchdog in CPLD Ctrl-1: %s\n",
			err)
	}

	if flag.ByName["-watchdog"] {
		c1 |= 0x3
		err = ioport.Outb(0x604, c1)
		if err != nil {
			return fmt.Errorf("Error setting watchdog in CPLD Ctrl-1: %s\n",
				err)
		}
	}
	c1 = (c1 & 0x7f) | byte(req<<7)
	err = ioport.Outb(0x604, c1)
	if err != nil {
		return fmt.Errorf("Error selecting QSPI in CPLD Ctrl-1: %s\n",
			err)
	}

	fmt.Printf("Selected QSPI%d\n", req)

	return nil
}
