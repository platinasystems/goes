// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package core provides a command collection similar to that provided by the
// debian coreutils package.
package core

import (
	"github.com/platinasystems/go/goes/core/cat"
	"github.com/platinasystems/go/goes/core/chmod"
	"github.com/platinasystems/go/goes/core/cp"
	"github.com/platinasystems/go/goes/core/echo"
	"github.com/platinasystems/go/goes/core/exec"
	"github.com/platinasystems/go/goes/core/kill"
	"github.com/platinasystems/go/goes/core/ln"
	"github.com/platinasystems/go/goes/core/log"
	"github.com/platinasystems/go/goes/core/ls"
	"github.com/platinasystems/go/goes/core/mkdir"
	"github.com/platinasystems/go/goes/core/ps"
	"github.com/platinasystems/go/goes/core/pwd"
	"github.com/platinasystems/go/goes/core/rm"
	"github.com/platinasystems/go/goes/core/sleep"
	"github.com/platinasystems/go/goes/core/stty"
	"github.com/platinasystems/go/goes/core/sync"
	"github.com/platinasystems/go/goes/core/toggle"
)

func New() []interface{} {
	return []interface{}{
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
		toggle.New(),
	}
}
