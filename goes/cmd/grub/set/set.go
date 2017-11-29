// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package set

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

type Command struct {
	g *goes.Goes
}

func (c Command) String() string { return "set" }

func (c Command) Usage() string {
	return "NOP"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "NOP",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: Man,
	}
}

const Man = "NOP command for script compatibility\n"

func (Command) Kind() cmd.Kind { return cmd.DontFork }

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c Command) Main(args ...string) error {
	s := strings.SplitN(args[0], "=", 2)
	if len(s) != 2 {
		return fmt.Errorf("unexpected %v\n", args)
	}
	if c.g.EnvMap == nil {
		c.g.EnvMap = make(map[string]string)
	}
	c.g.EnvMap[s[0]] = s[1]
	return nil
}
