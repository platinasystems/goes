// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package thencmd

import (
	"fmt"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

var errMissingIf = fmt.Errorf("then: missing if")

type Command struct {
	g *goes.Goes
}

func (*Command) String() string { return "then" }

func (*Command) Usage() string {
	return "if COMMAND ; then COMMAND else COMMAND endif"
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "conditionally execute commands",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Tests conditions and returns zero or non-zero exit status
`,
	}
}

func (*Command) Kind() cmd.Kind { return cmd.DontFork | cmd.Conditional }

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c *Command) Main(args ...string) error {
	if len(c.g.Blocks) == 0 {
		return errMissingIf
	}

	if c.g.Blocks[len(c.g.Blocks)-1] == goes.BlockIfNotTaken {
		return nil
	}
	if c.g.Blocks[len(c.g.Blocks)-1] != goes.BlockIf {
		return errMissingIf
	}
	if c.g.Status != nil {
		c.g.Blocks[len(c.g.Blocks)-1] = goes.BlockIfThenNotTaken
	} else {
		c.g.Blocks[len(c.g.Blocks)-1] = goes.BlockIfThenTaken
	}
	if len(args) == 0 {
		return nil
	}
	return c.g.Main(args...)
}
