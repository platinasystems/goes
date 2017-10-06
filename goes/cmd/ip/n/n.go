// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package n

import (
	"fmt"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/ip/internal/netns"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Apropos = "run ip command in namespace"
	Man     = `
SEE ALSO
	man ip || ip -man`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New(s string) *Command { return &Command{name: s} }

type Command struct {
	name string
	g    *goes.Goes
}

func (*Command) Apropos() lang.Alt   { return apropos }
func (c *Command) Goes(g *goes.Goes) { c.g = g }
func (*Command) Man() lang.Alt       { return man }
func (c *Command) String() string    { return c.name }
func (c *Command) Usage() string {
	return fmt.Sprintf("ip %s NAME OBJECT [ COMMAND [ ARGS ]...]",
		c.name)
}

func (c *Command) Main(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing NAME")
	}
	if err := netns.Switch(args[0]); err != nil {
		return err
	}
	return c.g.Main(args[1:]...)
}

func (c *Command) Complete(args ...string) (list []string) {
	switch len(args) {
	case 0:
		list = netns.CompleteName("")
	case 1:
		list = netns.CompleteName(args[0])
	default:
		list = c.g.Complete(args[1:]...)
	}
	return
}
