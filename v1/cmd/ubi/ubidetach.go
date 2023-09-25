// Copyright © 2019-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ubi

import (
	"fmt"
	"strconv"

	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/lang"

	"github.com/platinasystems/ubi"
)

type DetachCommand struct{}

func (DetachCommand) String() string { return "ubidetach" }

func (DetachCommand) Usage() string { return "ubidetach -d [UBI-DEV]" }

func (DetachCommand) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "Detach a UBI partition from a MTD device",
	}
}

func (DetachCommand) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	The ubidetach command is used to deassociate a MTD device from
	the UBI subsystem.

OPTIONS
	-d [UNIT]
		Specifies the unit number of the UBI device to detach.

EXAMPLES
	ubidetach -d 0
		Detach /dev/ubi0.`,
	}
}

func (c DetachCommand) Main(args ...string) (err error) {
	parm, args := parms.New(args, "-d")

	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	unit := 0

	if us := parm.ByName["-d"]; us != "" {
		unit, err = strconv.Atoi(us)
		if err != nil {
			return fmt.Errorf("Error parsing %s: %s\n", us, err)
		}
	}

	err = ubi.Detach(int32(unit))
	return err
}
