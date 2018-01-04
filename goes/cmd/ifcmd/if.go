// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ifcmd

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

type Command struct {
	g *goes.Goes
}

func (*Command) String() string { return "if" }

func (*Command) Usage() string {
	return "if COMMAND ; then COMMAND else COMMAND endif"
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "conditional command",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Conditionally executes statements in a script`,
	}
}

func (*Command) Kind() cmd.Kind { return cmd.DontFork | cmd.Conditional }

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c *Command) Main(args ...string) error {
	if c.g.NotTaken() {
		c.g.Blocks = append(c.g.Blocks, goes.BlockIfNotTaken)
		return nil
	}
	c.g.Blocks = append(c.g.Blocks, goes.BlockIf)
	if len(args) == 0 {
		return nil
	}
	return c.g.Main(args...)
}
