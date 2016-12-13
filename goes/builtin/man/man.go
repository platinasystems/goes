// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package man

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/goes"
)

const Name = "man"

type cmd goes.ByName

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.Builtin }
func (*cmd) String() string  { return "man" }
func (*cmd) Usage() string   { return "man COMMAND...\nCOMMAND -man" }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Complete(args ...string) []string {
	return goes.ByName(*c).Complete(args...)
}

func (c *cmd) Main(args ...string) error {
	n := len(args)
	if n == 0 {
		return fmt.Errorf("COMMAND: missing")
	}
	for i, arg := range args {
		var man string
		g := goes.ByName(*c)[arg]
		if g == nil {
			return fmt.Errorf("%s: not found", arg)
		}
		for _, lang := range []string{
			os.Getenv("LANG"),
			goes.Lang,
			goes.DefaultLang,
		} {
			man = g.Man[lang]
			if len(man) > 0 {
				break
			}
		}
		if len(man) == 0 {
			man = fmt.Sprint(arg, ": has no man")
		}
		fmt.Println(man)
		if n > 1 && i < n-1 {
			fmt.Println()
		}
	}
	return nil
}

func (*cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print command documentation",
	}
}
