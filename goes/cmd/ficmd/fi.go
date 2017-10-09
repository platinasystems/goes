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

const (
	Name    = "fi"
	Apropos = "fi"
	Usage   = "fi"
	Man     = `
DESCRIPTION
	Terminates an if block
`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() cmd.Cmd {
	return cmd.Cmd(new(Command))
}

type Command struct {
	g *goes.Goes
}

func (*Command) Apropos() lang.Alt   { return apropos }
func (*Command) Man() lang.Alt       { return man }
func (*Command) String() string      { return Name }
func (*Command) Usage() string       { return Usage }
func (*Command) Kind() cmd.Kind      { return cmd.DontFork | cmd.Conditional }
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
