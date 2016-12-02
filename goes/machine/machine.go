// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package machine provides machine admin commands and daemons.
package machine

import (
	"github.com/platinasystems/go/goes/machine/diag"
	"github.com/platinasystems/go/goes/machine/gpio"
	"github.com/platinasystems/go/goes/machine/i2c"
	"github.com/platinasystems/go/goes/machine/install"
	"github.com/platinasystems/go/goes/machine/reload"
	"github.com/platinasystems/go/goes/machine/restart"
	"github.com/platinasystems/go/goes/machine/slashinit"
	"github.com/platinasystems/go/goes/machine/start"
	"github.com/platinasystems/go/goes/machine/stop"
	"github.com/platinasystems/go/goes/machine/uninstall"
	"github.com/platinasystems/go/goes/machine/uptimed"
)

func New() []interface{} {
	return []interface{}{
		i2c.New(),
		gpio.New(),
		diag.New(),
		install.New(),
		reload.New(),
		restart.New(),
		start.New(),
		stop.New(),
		slashinit.New(),
		uninstall.New(),
		uptimed.New(),
	}
}
