// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package exit

import (
	"os"
	"strconv"

	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "exit" }

func (Command) Usage() string { return "exit [N]" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "exit the shell",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Exit the shell, returning a status of N, if given, or 0 otherwise.`,
	}
}

func (Command) Kind() cmd.Kind { return cmd.DontFork | cmd.CantPipe }

func (Command) Main(args ...string) error {
	var ecode int
	if len(args) != 0 {
		i64, err := strconv.ParseInt(args[0], 0, 0)
		if err != nil {
			return err
		}
		ecode = int(i64)
	}
	os.Exit(ecode)
	return nil
}
