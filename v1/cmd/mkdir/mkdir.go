// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mkdir

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/lang"
)

const DefaultMode = 0755

type Command struct{}

func (Command) String() string { return "mkdir" }

func (Command) Usage() string {
	return "mkdir [OPTION]... DIRECTORY..."
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "make directories",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Create the DIRECTORY(ies), if they do not already exist.

OPTIONS
	-m MODE
		set mode (as in chmod) to the given octal value;
		default, 0777 & umask()

	-p
		don't print error if existing and make parent directories
		as needed

	-v
		print a message for each created directory`,
	}
}

func (Command) Main(args ...string) error {
	var perm os.FileMode = DefaultMode

	flag, args := flags.New(args, "-p", "-v")
	parm, args := parms.New(args, "-m")

	if len(parm.ByName["-m"]) == 0 {
		parm.ByName["-m"] = "0755"
	}

	mode, err := strconv.ParseUint(parm.ByName["-m"], 8, 64)
	if err != nil {
		return err
	}

	if mode != DefaultMode {
		old := syscall.Umask(0)
		defer syscall.Umask(old)
		perm = os.FileMode(mode)
	}

	f := os.Mkdir
	if flag.ByName["-p"] {
		f = os.MkdirAll
	}

	for _, dn := range args {
		if err := f(dn, perm); err != nil {
			return err
		}
		if flag.ByName["-v"] {
			fmt.Printf("mkdir: created directory ‘%s’\n", dn)
		}
	}
	return nil
}
