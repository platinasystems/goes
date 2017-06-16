// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package complete

import (
	"fmt"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "complete"
	Apropos = "tab to complete command argument"
	Usage   = `
	complete COMMAND [ARGS]...
	COMMAND -complete [ARGS]...`
	Man = `
DESCRIPTION
	This may be used for bash completion of goes commands like this.

	_goes() {
		COMPREPLY=($(goes complete ${COMP_WORDS[@]}))
		return 0
	}
	complete -F _goes goes`
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
	for _, s := range c.g.Complete(args...) {
		fmt.Println(s)
	}
	return nil
}
