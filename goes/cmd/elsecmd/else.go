// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package elsecmd

import (
	"fmt"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

type Command struct {
	g *goes.Goes
}

func (*Command) String() string { return "else" }

func (*Command) Usage() string { return "else COMMAND" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "if COMMAND ; then COMMAND else COMMAND endelse",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Conditionally executes statements in a script
`,
	}
}

func (*Command) Kind() cmd.Kind { return cmd.DontFork | cmd.Conditional }

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c *Command) Main(args ...string) error {
	if len(c.g.Blocks) == 0 {
		return fmt.Errorf("%v: missing if", args)
	}

	if c.g.Blocks[len(c.g.Blocks)-1] == goes.BlockIfNotTaken {
		return nil
	}
	if c.g.Blocks[len(c.g.Blocks)-1] != goes.BlockIfThenTaken &&
		c.g.Blocks[len(c.g.Blocks)-1] != goes.BlockIfThenNotTaken {
		return fmt.Errorf("%v: missing then", args)
	}

	if c.g.Blocks[len(c.g.Blocks)-1] == goes.BlockIfThenTaken {
		c.g.Blocks[len(c.g.Blocks)-1] = goes.BlockIfElseNotTaken
	} else {
		c.g.Blocks[len(c.g.Blocks)-1] = goes.BlockIfElseTaken
	}
	if len(args) == 0 {
		return nil
	}
	return c.g.Main(args...)
}
