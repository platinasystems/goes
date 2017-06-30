// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package source

import (
	"fmt"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
)

const (
	Name    = "source"
	Apropos = "import command script"
	Usage   = "source [-x] FILE"
	Man     = `
DESCRIPTION
	This is equivalent to 'cli [-x] URL'.`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() *Command { return new(Command) }

type Command struct {
	g *goes.Goes
}

func (*Command) Apropos() lang.Alt   { return apropos }
func (c *Command) Goes(g *goes.Goes) { c.g = g }
func (*Command) Kind() cmd.Kind      { return cmd.DontFork | cmd.CantPipe }

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
	return c.g.Main(args...)
}

func (*Command) Man() lang.Alt  { return man }
func (*Command) String() string { return Name }
func (*Command) Usage() string  { return Usage }
