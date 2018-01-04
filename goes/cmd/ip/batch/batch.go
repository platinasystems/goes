// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package batch

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
)

type Command struct {
	g *goes.Goes
}

func (*Command) String() string { return "-batch" }

func (*Command) Usage() string {
	return `ip [-n NAMESPACE] -batch  [ -x | -f ] [ - | FILE ]`
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "ip commands from file or stdin",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
SEE ALSO
	man ip || ip -man`,
	}
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c *Command) Main(args ...string) error {
	return c.g.Main(args...)
}

func (Command) Complete(args ...string) (list []string) {
	var larg, llarg string
	n := len(args)
	if n > 0 {
		larg = args[n-1]
	}
	if n > 1 {
		llarg = args[n-2]
	}
	if method, found := options.CompleteParmValue[llarg]; found {
		list = method(larg)
	} else {
		for _, name := range append(options.CompleteOptNames,
			"-x",
			"-f",
			"-") {
			if len(larg) == 0 || strings.HasPrefix(name, larg) {
				list = append(list, name)
			}
		}
	}
	if len(list) == 0 {
		list, _ = filepath.Glob(larg + "*")
	}
	if len(list) > 0 {
		sort.Strings(list)
	}
	return
}
