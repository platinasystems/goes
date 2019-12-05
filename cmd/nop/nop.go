// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nop

import (
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/lang"
)

type Command struct {
	C string
}

func (c Command) String() string {
	if c.C == "" {
		return "nop"
	}
	return c.C
}

func (c Command) Usage() string { return c.String() + " ..." }

func (c Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "do nothing",
	}
}

func (c Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	The ` + c.String() + ` command does nothing. It is intended for use in
	scripts.`,
	}
}

func (Command) Kind() cmd.Kind { return cmd.DontFork }

func (Command) Main(args ...string) error {
	return nil
}
