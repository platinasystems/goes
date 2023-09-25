// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"fmt"

	"github.com/platinasystems/goes/lang"
)

type aproposer interface {
	Apropos() lang.Alt
}

func (g *Goes) Apropos() lang.Alt {
	apropos := g.APROPOS
	if apropos == nil {
		apropos = lang.Alt{
			lang.EnUS: "a golang busybox",
		}
	}
	return apropos
}

func (g *Goes) apropos(args ...string) error {
	pad := func(n int) {
		if n < 0 {
			fmt.Print("\n\t\t")
		} else {
			fmt.Print("                "[:n])
		}
	}
	if len(args) == 0 {
		args = g.Names()
	}
	for i, name := range args {
		if len(name) == 0 {
			continue
		}
		if cmd, found := g.ByName[name]; found {
			fmt.Print(name)
			pad(16 - len(name))
			fmt.Println(cmd.Apropos())
		} else if i == 0 {
			return fmt.Errorf("%s: not found", name)
		}
	}
	return nil
}
