// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package machine provides machine admin commands and daemons.
package machine

import (
	"github.com/platinasystems/go/commands/machine/diag"
	"github.com/platinasystems/go/commands/machine/gpio"
	"github.com/platinasystems/go/commands/machine/i2c"
	"github.com/platinasystems/go/commands/machine/install"
	"github.com/platinasystems/go/commands/machine/machined"
	"github.com/platinasystems/go/commands/machine/reload"
	"github.com/platinasystems/go/commands/machine/restart"
	"github.com/platinasystems/go/commands/machine/slashinit"
	"github.com/platinasystems/go/commands/machine/start"
	"github.com/platinasystems/go/commands/machine/stop"
	"github.com/platinasystems/go/commands/machine/uninstall"
)

func New() []interface{} {
	return []interface{}{
		i2c.New(),
		gpio.New(),
		diag.New(),
		install.New(),
		machined.New(),
		reload.New(),
		restart.New(),
		start.New(),
		stop.New(),
		slashinit.New(),
		uninstall.New(),
	}
}
