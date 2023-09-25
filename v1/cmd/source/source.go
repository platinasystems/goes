// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package source

import (
	"fmt"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/lang"
)

type Command struct {
	g *goes.Goes
}

func (*Command) String() string { return "source" }

func (*Command) Usage() string {
	return "source [-x] FILE"
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "import command script",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	This is equivalent to 'cli [-x] URL'.`,
	}
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (*Command) Kind() cmd.Kind { return cmd.DontFork | cmd.CantPipe }

func (c *Command) Main(args ...string) error {
	flag, args := flags.New(args, "-x")
	if len(args) == 0 {
		return fmt.Errorf("FILE: missing")
	}
	if len(args) > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	if flag.ByName["-x"] {
		args = []string{"cli", "-x", args[0]}
	} else {
		args = []string{"cli", args[0]}
	}
	c.g.Catline = nil // Reset the input source
	return c.g.Main(args...)
}
