// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package elsecmd

import (
	"errors"

	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "else" }

func (Command) Usage() string { return "else COMMAND" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "if COMMAND ; then COMMAND else COMMAND endelse",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Conditionally executes statements in a script
`,
	}
}

func (c Command) Main(args ...string) error {
	return errors.New("missing if")
}
