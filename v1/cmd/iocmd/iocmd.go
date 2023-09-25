// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build linux,amd64

package iocmd

import (
	"fmt"
	"strconv"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/ioport"
)

type Command struct{}

func (Command) String() string { return "io" }

func (Command) Usage() string {
	return "io [[-r] | -w] IO-ADDRESS [-D DATA] [-m MODE]"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "read/write the CPU's I/O ports",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	This command reads and writes the CPU's I/O ports.
	  -r to read from ioport, default
	  -w to write from ioport
	     IO-ADDRESS is a hex value
	  -D DATA is a hex value`,
	}
}

func (Command) Main(args ...string) (err error) {
	flag, args := flags.New(args, "-r", "-w")
	parm, args := parms.New(args, "-D")
	if len(args) == 0 {
		return fmt.Errorf("IO-ADDRESS: missing")
	}
	if parm.ByName["-D"] == "" {
		parm.ByName["-D"] = "0x0"
	}

	var a, d uint64

	if a, err = strconv.ParseUint(args[0], 0, 16); err != nil {
		return fmt.Errorf("%s: %v", args[0], err)
	}
	if d, err = strconv.ParseUint(parm.ByName["-D"], 0, 8); err != nil {
		return fmt.Errorf("%s: %v", parm.ByName["-D"], err)
	}

	if flag.ByName["-w"] {
		if err = ioport.Outb(uint16(a), byte(d)); err != nil {
			return err
		}
	} else {
		b := byte(0)
		if b, err = ioport.Inb(uint16(a)); err != nil {
			return err
		}
		fmt.Printf("%x: %x\n", a, b)
	}
	return nil
}
