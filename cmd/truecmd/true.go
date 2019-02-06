// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package truecmd

import "github.com/platinasystems/goes/lang"

type Command struct{}

func (Command) String() string { return "true" }

func (Command) Usage() string { return "true" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "Be successful not matter what",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Be successful no matter what!`,
	}
}

func (Command) Main(_ ...string) error {
	return nil
}
