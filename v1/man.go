// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"fmt"
	"strings"

	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/lang"
)

type maner interface {
	Man() lang.Alt
}

var section = struct {
	name, synopsis lang.Alt
}{
	name: lang.Alt{
		lang.EnUS: "NAME",
	},
	synopsis: lang.Alt{
		lang.EnUS: "SYNOPSIS",
	},
}

func (g *Goes) Man() lang.Alt {
	man := g.MAN
	if man == nil {
		man = lang.Alt{
			lang.EnUS: `
OPTIONS
	-d	debug block handling
	-x	print command trace
	-f	don't terminate script on error
	-	execute standard input script
	SCRIPT	execute named script file

SEE ALSO
	goes apropos [COMMAND], goes man COMMAND`,
		}

	}
	return man
}

func (g *Goes) man(args ...string) error {
	var cmds []cmd.Cmd
	for i, arg := range args {
		v := g.ByName[arg]
		if v == nil {
			if i == 0 {
				return fmt.Errorf("%s: not found", arg)
			}
			break
		}
		cmds = append(cmds, v)
	}
	if len(cmds) == 0 {
		cmds = []cmd.Cmd{g}
	}
	for i, v := range cmds {
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
