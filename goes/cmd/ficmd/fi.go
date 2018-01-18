// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ficmd

import (
	"errors"

	"github.com/platinasystems/go/goes/lang"
)

type Command struct{}

func (Command) String() string { return "fi" }

func (Command) Usage() string { return "fi" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "end of if command block",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Terminates an if block
`,
	}
}

func (c Command) Main(args ...string) error {
	return errors.New("missing if")
}
