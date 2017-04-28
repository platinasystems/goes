// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package apropos

import (
	"fmt"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "apropos"
	Apropos = "print a short command description"
	Usage   = `
	apropos COMMAND...
	COMMAND -apropos`
)

type Interface interface {
	Apropos() lang.Alt
	ByName(goes.ByName)
	Complete(...string) []string
	Kind() goes.Kind
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return new(cmd) }

type cmd goes.ByName

func (*cmd) Apropos() lang.Alt { return apropos }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Complete(args ...string) []string {
	var prefix string
	if len(args) > 0 {
		prefix = args[len(args)-1]
	}
	return goes.ByName(*c).Complete(prefix)
}

func (*cmd) Kind() goes.Kind { return goes.DontFork }

func (c *cmd) Main(args ...string) error {
	pad := func(n int) {
		if n < 0 {
			fmt.Print("\n\t\t")
		} else {
			fmt.Print("                "[:n])
		}
	}
	if len(args) == 0 {
		for _, k := range goes.ByName(*c).Complete("") {
			apropos := goes.ByName(*c)[k].Apropos
			if apropos != nil {
				fmt.Print(k)
				pad(16 - len(k))
				fmt.Println(apropos)
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
			fmt.Print(k)
			pad(16 - len(k))
			fmt.Println(g.Apropos)
		}
	}
	return nil
}

func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
