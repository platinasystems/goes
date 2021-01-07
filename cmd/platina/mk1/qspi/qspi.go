// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package qspi

import (
	"fmt"
	"strconv"

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
	The qspi command reports or sets the active QSPI device.`,
	}
}

func (c Command) Main(args ...string) (err error) {
	sel := 0
	v, err := ioport.Inb(0x604)
	if err != nil {
		return fmt.Errorf("Error in Inb(0x604): %s", err)
	}
	if v&0x80 != 0 {
		sel = 1
	}

	if len(args) == 0 {
		fmt.Printf("QSPI%d is selected\n", sel)
		return nil
	}

	if len(args) > 0 {
		sel, err = strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("Error parsing %s: %s", args[0], err)
		}
		args = args[1:]
		if len(args) > 0 {
			return fmt.Errorf("%v: unexpected", args)
		}
		if sel < 0 || sel > 1 {
			return fmt.Errorf("Invalid unit number %d", sel)
		}
	}

	v = (v & 0x7f) | byte(sel<<7)

	err = ioport.Outb(0x604, v)
	if err != nil {
		return fmt.Errorf("Error in outb(0x604): %s\n",
			err)
	}

	fmt.Printf("Selected QSPI%d\n", sel)

	return nil
}
