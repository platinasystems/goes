// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package all

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/ip/internal/netns"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Apropos = "run ip command in all namespaces"
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
	return fmt.Sprintf("ip %s OBJECT [ COMMAND [ ARGS ]...]",
		c.name)
}

func (c *Command) Main(args ...string) error {
	for _, name := range netns.List() {
		cmd := c.g.Fork(append([]string{"-netns", name}, args...)...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("%v: %v", cmd.Args, err)
		}
	}
	return c.g.Main(args...)
}

func (c *Command) Complete(args ...string) (list []string) {
	return c.g.Complete(args...)
}
