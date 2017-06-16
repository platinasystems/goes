// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package helpers

import (
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/cmd/apropos"
	"github.com/platinasystems/go/goes/cmd/complete"
	"github.com/platinasystems/go/goes/cmd/help"
	"github.com/platinasystems/go/goes/cmd/man"
	"github.com/platinasystems/go/goes/cmd/usage"
)

func New() []cmd.Cmd {
	return []cmd.Cmd{
		apropos.New(),
		complete.New(),
		help.New(),
		man.New(),
		usage.New(),
	}
}
