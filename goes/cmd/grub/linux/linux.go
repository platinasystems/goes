// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package linux

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

type Command struct {
	Kern string
	Cmd  []string
}

func (c Command) String() string { return "linux" }

func (c Command) Usage() string {
	return "NOP"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "NOP",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: Man,
	}
}

const Man = "Set the kernel and command line"

func (Command) Kind() cmd.Kind { return cmd.DontFork }

func (c *Command) Main(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("%s: missing arguments\n", c)
	}
	c.Kern = args[0]
	if len(args) > 1 {
		c.Cmd = args[1:]
	}
	return nil
}
