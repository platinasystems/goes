// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cmd

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/apropos"
	"github.com/platinasystems/go/goes/cmd/complete"
	"github.com/platinasystems/go/goes/cmd/help"
	"github.com/platinasystems/go/goes/cmd/man"
	"github.com/platinasystems/go/goes/cmd/show_packages"
	"github.com/platinasystems/go/goes/cmd/usage"
)

// Returns a goes.ByName with the given plus these flag initiated commands:
//	apropos, complete, help, license, man, patents, usage, version
func New(cmd ...goes.Cmd) goes.ByName {
	return goes.New().Plot(
		apropos.New(),
		complete.New(),
		help.New(),
		man.New(),
		show_packages.New(),
		usage.New(),
	).Plot(cmd...)
}
