// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ifcmd

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "if"
	Apropos = "if COMMAND ; then COMMAND else COMMAND endif"
	Usage   = "if COMMAND"
	Man     = `
DESCRIPTION
	Conditionally executes statements in a script
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
