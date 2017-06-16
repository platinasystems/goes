// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package man

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "man"
	Apropos = "print command documentation"
	Usage   = `
	man COMMAND...
	COMMAND -man`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	section = struct {
		name, synopsis lang.Alt
	}{
		name: lang.Alt{
			lang.EnUS: "NAME",
		},
		synopsis: lang.Alt{
			lang.EnUS: "SYNOPSIS",
		},
	}
)

func New() *Command { return new(Command) }

type Command struct {
	g *goes.Goes
}

type maner interface {
	Man() lang.Alt
}

func (*Command) Apropos() lang.Alt   { return apropos }
func (c *Command) Goes(g *goes.Goes) { c.g = g }
func (*Command) String() string      { return Name }
func (*Command) Usage() string       { return Usage }

func (c *Command) Main(args ...string) error {
	var vs []cmd.Cmd
	if len(args) == 0 {
		vs = []cmd.Cmd{cmd.Cmd(c.g)}
	} else {
		for i, arg := range args {
			v := c.g.ByName(arg)
			if v == nil {
				if i == 0 {
					return fmt.Errorf("%s: not found", arg)
				}
				break
			}
			vs = append(vs, v)
		}
	}
	for i, v := range vs {
		if i > 0 {
			fmt.Println()
		}
		fmt.Print(section.name, "\n\t", v, " - ",
			v.Apropos(), "\n\n", section.synopsis, "\n\t",
			strings.TrimSpace(v.Usage()), "\n")
		if method, found := v.(maner); found {
			man := method.Man().String()
			if !strings.HasPrefix(man, "\n") {
				fmt.Println()
			}
			fmt.Print(man)
			if !strings.HasSuffix(man, "\n") {
				fmt.Println()
			}
		}
	}
	return nil
}
