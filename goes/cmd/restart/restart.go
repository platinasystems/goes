// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package restart

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

type Command struct {
	g *goes.Goes
}

func (*Command) String() string { return "restart" }

func (*Command) Usage() string {
	return "restart [STOP, STOP, and REDISD OPTIONS]..."
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "stop, then start this goes machine",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Run the goes machine stop then start commands.

SEE ALSO
	start, stop, and redisd`,
	}
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c *Command) Main(args ...string) error {
	err := c.g.Main(append([]string{"stop"}, args...)...)
	if err != nil {
		return err
	}
	return c.g.Main(append([]string{"start"}, args...)...)
}
