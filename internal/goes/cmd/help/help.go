// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package help

import (
	"fmt"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/goes/lang"
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

type Interface interface {
	Apropos() lang.Alt
	ByName(goes.ByName)
	Complete(...string) []string
	Kind() goes.Kind
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return new(cmd) }

type cmd goes.ByName

func (*cmd) Apropos() lang.Alt { return apropos }

func (c *cmd) Complete(args ...string) []string {
	var prefix string
	if len(args) > 0 {
		prefix = args[len(args)-1]
	}
	return goes.ByName(*c).Complete(prefix)
}

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (*cmd) Kind() goes.Kind { return goes.DontFork }

func (c *cmd) Main(args ...string) error {
	if len(args) == 0 {
		for _, k := range goes.ByName(*c).Complete("") {
			g := goes.ByName(*c)[k]
			apropos := g.Apropos.String()
			if len(apropos) > 0 {
				format := "%-15s %s\n"
				if len(k) >= 16 {
					format = "%s\n\t\t%s\n"
				}
				fmt.Printf(format, k, apropos)
			}
		}
	} else {
		g := goes.ByName(*c)[args[0]]
		if g == nil {
			return fmt.Errorf("%s: not found", args[0])
		}
		if g.Help != nil {
			fmt.Println(g.Help(args[1:]...))
		} else {
			fmt.Println(g.Usage)
		}
	}
	return nil
}

func (*cmd) Man() lang.Alt  { return man }
func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
