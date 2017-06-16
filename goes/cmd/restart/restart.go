// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package restart

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "restart"
	Apropos = "stop, then start this goes machine"
	Usage   = "restart [STOP, STOP, and REDISD OPTIONS]..."
	Man     = `
DESCRIPTION
	Run the goes machine stop then start commands.

SEE ALSO
	start, stop, and redisd`
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

func (*Command) Apropos() lang.Alt { return apropos }

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c *Command) Main(args ...string) error {
	err := c.g.Main(append([]string{"stop"}, args...)...)
	if err != nil {
		return err
	}
	return c.g.Main(append([]string{"start"}, args...)...)
}

func (*Command) Man() lang.Alt  { return man }
func (*Command) String() string { return Name }
func (*Command) Usage() string  { return Usage }
