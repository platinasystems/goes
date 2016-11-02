// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package initutils provides /init, install, start, stop, restart, and reload
// commands.
package initutils

import (
	"github.com/platinasystems/go/initutils/install"
	"github.com/platinasystems/go/initutils/reload"
	"github.com/platinasystems/go/initutils/restart"
	"github.com/platinasystems/go/initutils/slashinit"
	"github.com/platinasystems/go/initutils/start"
	"github.com/platinasystems/go/initutils/stop"
	"github.com/platinasystems/go/initutils/uninstall"
)

func New() []interface{} {
	return []interface{}{
		install.New(),
		reload.New(),
		restart.New(),
		start.New(),
		stop.New(),
		slashinit.New(),
		uninstall.New(),
	}
}
