// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package export

import (
	"fmt"
	"os"
	"strings"

	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "export" }

func (Command) Usage() string { return "export [NAME[=VALUE]]..." }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "set process configuration",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Configure the named process environment parameter.

	If no VALUE is given, NAME is reset.

	If no NAMES are supplied, a list of names of all exported variables
	is printed.`,
	}
}

func (Command) Kind() cmd.Kind { return cmd.DontFork | cmd.CantPipe }

func (Command) Main(args ...string) error {
	if len(args) == 0 {
		for _, nv := range os.Environ() {
			fmt.Println(nv)
		}
		return nil
	}
	for _, arg := range args {
		eq := strings.Index(arg, "=")
		if eq < 0 {
			if err := os.Unsetenv(arg); err != nil {
				return err
			}
		} else if err := os.Setenv(arg[:eq], arg[eq+1:]); err != nil {
			return err
		}
	}
	return nil
}
