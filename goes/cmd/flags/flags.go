// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package flags

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/apropos"
	"github.com/platinasystems/go/goes/cmd/complete"
	"github.com/platinasystems/go/goes/cmd/help"
	"github.com/platinasystems/go/goes/cmd/license"
	"github.com/platinasystems/go/goes/cmd/man"
	"github.com/platinasystems/go/goes/cmd/patents"
	"github.com/platinasystems/go/goes/cmd/usage"
	"github.com/platinasystems/go/goes/cmd/version"
)

// Returns these flag initiated commands:
//	apropos, complete, help, license, man, patents, usage, version
func New(c ...goes.Cmd) []goes.Cmd {
	return []goes.Cmd{
		apropos.New(),
		complete.New(),
		help.New(),
		license.New(),
		man.New(),
		patents.New(),
		usage.New(),
		version.New(),
	}
}
