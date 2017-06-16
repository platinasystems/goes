// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package usage

import (
	"fmt"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "usage"
	Apropos = "print a command synopsis"
	Usage   = `
	usage COMMAND
	COMMAND -usage`
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
	v := cmd.Cmd(c.g)
	if len(args) > 0 {
		v = c.g.ByName(args[0])
		if v == nil {
			return fmt.Errorf("%s: not found", args[0])
		}
	}
	fmt.Println(goes.Usage(v))
	return nil
}
