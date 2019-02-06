// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cd

import (
	"fmt"
	"os"

	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/lang"
)

type Command struct {
	last string
}

func (*Command) String() string { return "cd" }

func (*Command) Usage() string { return "cd [- | DIRECTORY]" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "change the current directory",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Change the working directory to the given name or the last one if '-'.
	`,
	}

}

func (*Command) Kind() cmd.Kind { return cmd.DontFork | cmd.CantPipe }

func (cd *Command) Main(args ...string) error {
	var dir string

	if len(args) > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	t, err := os.Getwd()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		dir = os.Getenv("HOME")
		if len(dir) == 0 {
			dir = "/root"
		}
	} else if args[0] == "-" {
		if len(cd.last) > 0 {
			dir = cd.last
		}
	} else {
		dir = args[0]
	}
	if len(dir) > 0 {
		err := os.Chdir(dir)
		if err == nil {
			cd.last = t
		}
		return err
	}
	return nil
}
