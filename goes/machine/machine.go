// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package machine provides machine admin commands and daemons.
package machine

import (
	"github.com/platinasystems/go/goes/machine/diag"
	"github.com/platinasystems/go/goes/machine/gpio"
	"github.com/platinasystems/go/goes/machine/i2c"
	"github.com/platinasystems/go/goes/machine/install"
	"github.com/platinasystems/go/goes/machine/machined"
	"github.com/platinasystems/go/goes/machine/reload"
	"github.com/platinasystems/go/goes/machine/restart"
	"github.com/platinasystems/go/goes/machine/slashinit"
	"github.com/platinasystems/go/goes/machine/start"
	"github.com/platinasystems/go/goes/machine/stop"
	"github.com/platinasystems/go/goes/machine/uninstall"
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
