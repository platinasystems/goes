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

type AttachCommand struct{}

func (AttachCommand) String() string { return "ubiattach" }

func (AttachCommand) Usage() string { return "ubiattach [OPTIONS]" }

func (AttachCommand) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "attach a UBI partition to a MTD device",
	}
}

func (AttachCommand) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	The ubiattach command is used to associated a MTD device with
	the UBI subsystem. The underlying MTD device or partition must
	either contain a valid UBI partition, or be erased (in which case
	the UBI subsystem will format the partition as UBI).

OPTIONS
	-d [UNIT]
		Specifies the unit number of the associated UBI device

	-m [UNIT]
		Specifies the unit number of the associated MTD device

EXAMPLES
	ubiattach -d 0 -m 5
		Attaches /dev/mtd5 to /dev/ubi0`,
	}
}

func (c AttachCommand) Main(args ...string) (err error) {
	parm, args := parms.New(args, "-d", "-m")

	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	mtdUnit := 0
	ubiUnit := 0

	if us := parm.ByName["-d"]; us != "" {
		ubiUnit, err = strconv.Atoi(us)
		if err != nil {
			return fmt.Errorf("Error parsing %s: %s\n", us, err)
		}
	}

	if us := parm.ByName["-m"]; us != "" {
		mtdUnit, err = strconv.Atoi(us)
		if err != nil {
			return fmt.Errorf("Error parsing %s: %s\n", us, err)
		}
	}

	err = ubi.Attach(int32(ubiUnit), int32(mtdUnit), 0, 0)
	return err
}
