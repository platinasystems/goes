// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show_commands

import (
	"fmt"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "show-commands"
	Apropos = "list all commands and daemons"
	Usage   = "show-commands"
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
func (*Command) Kind() cmd.Kind      { return cmd.DontFork }

func (c *Command) Main(args ...string) error {
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	for _, name := range c.g.Names {
		v := c.g.ByName(name)
		k := cmd.WhatKind(v)
		switch {
		case k.IsDaemon():
			fmt.Println(v, "- daemon")
		case k.IsHidden():
			fmt.Println(v, "- hidden")
		default:
			fmt.Println(v)
		}
	}
	return nil
}

func (*Command) String() string { return Name }
func (*Command) Usage() string  { return Usage }
