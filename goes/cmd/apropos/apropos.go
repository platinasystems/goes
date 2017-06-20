// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package apropos

import (
	"fmt"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "apropos"
	Apropos = "print a short command description"
	Usage   = `
	apropos [COMMAND]...
	COMMAND -apropos`
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() *Command { return new(Command) }

type Command struct {
	g *goes.Goes
}

func (*Command) Apropos() lang.Alt   { return apropos }
func (c *Command) Goes(g *goes.Goes) { c.g = g }
func (*Command) String() string      { return Name }
func (*Command) Usage() string       { return Usage }

func (c *Command) Main(args ...string) error {
	pad := func(n int) {
		if n < 0 {
			fmt.Print("\n\t\t")
		} else {
			fmt.Print("                "[:n])
		}
	}
	if len(args) == 0 {
		args = c.g.Names
	}
	for i, name := range args {
		if len(name) == 0 {
			continue
		}
		v := c.g.ByName(name)
		if v == nil {
			if i == 0 {
				return fmt.Errorf("%s: not found", name)
			}
			return nil
		}
		fmt.Print(name)
		pad(16 - len(name))
		fmt.Println(v.Apropos())
	}
	return nil
}
