// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package apropos

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/goes"
)

const Name = "apropos"

type cmd goes.ByName

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.Builtin }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return "apropos COMMAND...\nCOMMAND -apropos" }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Complete(args ...string) []string {
	return goes.ByName(*c).Complete(args...)
}

func (c *cmd) Main(args ...string) error {
	printApropos := func(apropos map[string]string, key string) {
		format := "%-15s %s\n"
		if len(key) >= 16 {
			format = "%s\n\t\t%s\n"
		}
		for _, lang := range []string{
			os.Getenv("LANG"),
			goes.Lang,
			goes.DefaultLang,
		} {
			if s := apropos[lang]; len(s) > 0 {
				fmt.Printf(format, key, s)
				break
			}
		}
	}
	if len(args) == 0 {
		for _, k := range goes.ByName(*c).Keys() {
			apropos := goes.ByName(*c)[k].Apropos
			if apropos != nil {
				printApropos(apropos, k)
			}
		}
	} else {
		for _, k := range args {
			g := goes.ByName(*c)[k]
			if g == nil {
				return fmt.Errorf("%s: not found", k)
			}
			if g.Apropos == nil {
				return fmt.Errorf("%s: has no apropos", k)
			}
			printApropos(g.Apropos, k)
		}
	}
	return nil
}

func (*cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print a short command description",
	}
}

func (*cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	apropos - print a short command description

SYNOPSIS
	apropos [COMMAND]...

DESCRIPTION
	Print a short description of given or all COMMANDS.`,
	}
}
