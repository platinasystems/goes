// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package thencmd

import (
	"errors"

	"github.com/platinasystems/go/goes/lang"
)

type Command struct{}

func (Command) String() string { return "then" }

func (Command) Usage() string {
	return "if COMMAND ; then COMMAND else COMMAND endif"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "conditionally execute commands",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Tests conditions and returns zero or non-zero exit status
`,
	}
}

func (c Command) Main(args ...string) error {
	return errors.New("missing if")
}
