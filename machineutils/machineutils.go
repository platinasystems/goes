// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package initutils provides /init, install, start, stop, restart, and reload
// commands along with the machind daemon.
package machineutils

import (
	"github.com/platinasystems/go/machineutils/install"
	"github.com/platinasystems/go/machineutils/machined"
	"github.com/platinasystems/go/machineutils/reload"
	"github.com/platinasystems/go/machineutils/restart"
	"github.com/platinasystems/go/machineutils/slashinit"
	"github.com/platinasystems/go/machineutils/start"
	"github.com/platinasystems/go/machineutils/stop"
	"github.com/platinasystems/go/machineutils/uninstall"
)

func New() []interface{} {
	return []interface{}{
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
