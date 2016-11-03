// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package core provides a command collection similar to that provided by the
// debian coreutils package.
package core

import (
	"github.com/platinasystems/go/commands/core/bang"
	"github.com/platinasystems/go/commands/core/cat"
	"github.com/platinasystems/go/commands/core/chmod"
	"github.com/platinasystems/go/commands/core/cp"
	"github.com/platinasystems/go/commands/core/echo"
	"github.com/platinasystems/go/commands/core/exec"
	"github.com/platinasystems/go/commands/core/kill"
	"github.com/platinasystems/go/commands/core/ln"
	"github.com/platinasystems/go/commands/core/log"
	"github.com/platinasystems/go/commands/core/ls"
	"github.com/platinasystems/go/commands/core/mkdir"
	"github.com/platinasystems/go/commands/core/ps"
	"github.com/platinasystems/go/commands/core/pwd"
	"github.com/platinasystems/go/commands/core/rm"
	"github.com/platinasystems/go/commands/core/sleep"
	"github.com/platinasystems/go/commands/core/stty"
	"github.com/platinasystems/go/commands/core/sync"
)

func New() []interface{} {
	return []interface{}{
		bang.New(),
		cat.New(),
		chmod.New(),
		cp.New(),
		echo.New(),
		exec.New(),
		kill.New(),
		ln.New(),
		log.New(),
		ls.New(),
		mkdir.New(),
		ps.New(),
		pwd.New(),
		rm.New(),
		sleep.New(),
		stty.New(),
		sync.New(),
	}
}
