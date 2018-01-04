// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ficmd

import (
	"fmt"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

type Command struct {
	g *goes.Goes
}

func (*Command) String() string { return "fi" }

func (*Command) Usage() string { return "fi" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "end of if command block",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Terminates an if block
`,
	}
}

func (*Command) Kind() cmd.Kind { return cmd.DontFork | cmd.Conditional }

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c *Command) Main(args ...string) error {
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	if len(c.g.Blocks) == 0 {
		return fmt.Errorf("%v: missing if", args)
	}

	if c.g.Blocks[len(c.g.Blocks)-1] == goes.BlockIfNotTaken ||
		c.g.Blocks[len(c.g.Blocks)-1] == goes.BlockIfThenTaken ||
		c.g.Blocks[len(c.g.Blocks)-1] == goes.BlockIfThenNotTaken ||
		c.g.Blocks[len(c.g.Blocks)-1] == goes.BlockIfElseTaken ||
		c.g.Blocks[len(c.g.Blocks)-1] == goes.BlockIfElseNotTaken {
		c.g.Blocks = c.g.Blocks[:len(c.g.Blocks)-1]
		return nil
	}
	return fmt.Errorf("%v: missing then", args)
}
