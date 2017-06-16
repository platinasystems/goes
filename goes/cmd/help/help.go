// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package help

import (
	"fmt"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "help"
	Apropos = "print command guidance"
	Usage   = `
	help [COMMAND [ARGS]...]
	COMMAND -help [ARGS]...
	`
	Man = `
DESCRIPTION
	Print context sensitive command help, if available; otherwise, print
	its usage page.

	Print all available apropos if no COMMAND is given.`
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
func (*Command) Man() lang.Alt       { return man }
func (*Command) String() string      { return Name }
func (*Command) Usage() string       { return Usage }

func (c *Command) Main(args ...string) error {
	fmt.Println(c.g.Help(args...))
	return nil
}
