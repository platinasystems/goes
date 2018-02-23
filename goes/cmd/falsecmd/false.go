// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package falsecmd

import (
	"os"

	"github.com/platinasystems/go/goes/lang"
)

type Command struct{}

func (Command) String() string { return "false" }

func (Command) Usage() string { return "false" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "Fail regardless of our ability",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Fail, not matter what. This can not happen in the real world.`,
	}
}

func (Command) Main(_ ...string) error {
	os.Exit(1)
	return nil
}
