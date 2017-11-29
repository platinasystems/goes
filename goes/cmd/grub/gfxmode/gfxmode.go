// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package gfxmode

import (
	"github.com/platinasystems/go/goes/lang"
)

type Command struct{}

func (c Command) String() string { return "gfxmode" }

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

func (Command) Main(args ...string) error {
	return nil
}
