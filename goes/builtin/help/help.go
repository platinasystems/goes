// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package help

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/goes"
)

const Name = "help"

type cmd goes.ByName

func New() *cmd { return new(cmd) }

func (*cmd) String() string { return Name }
func (*cmd) Tag() string    { return "builtin" }

func (*cmd) Usage() string {
	return "help [COMMAND [ARGS]...]\nCOMMAND -help [ARGS]..."
}

func (c *cmd) Complete(args ...string) []string {
	return goes.ByName(*c).Complete(args...)
}

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Main(args ...string) error {
	if len(args) == 0 {
		for _, k := range goes.ByName(*c).Keys() {
			g := goes.ByName(*c)[k]
			for _, lang := range []string{
				os.Getenv("LANG"),
				goes.Lang,
				goes.DefaultLang,
			} {
				s := g.Apropos[lang]
				if len(s) > 0 {
					format := "%-15s %s\n"
					if len(k) >= 16 {
						format = "%s\n\t\t%s\n"
					}
					fmt.Printf(format, k, s)
					break
				}
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

func (*cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print command guidance",
	}
}

func (*cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	help - print a command guidance

SYNOPSIS
	help [COMMAND [ARGS]...]

DESCRIPTION
	Print context sensitive command help, if available; otherwise, print
	its usage page.

	Print all available apropos if no COMMAND is given.`,
	}
}
