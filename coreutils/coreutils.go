// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package coreutils provides a command collection similar to that provided by
// the so named debian package.
package coreutils

import (
	"github.com/platinasystems/go/coreutils/bang"
	"github.com/platinasystems/go/coreutils/cat"
	"github.com/platinasystems/go/coreutils/chmod"
	"github.com/platinasystems/go/coreutils/cp"
	"github.com/platinasystems/go/coreutils/echo"
	"github.com/platinasystems/go/coreutils/exec"
	"github.com/platinasystems/go/coreutils/kill"
	"github.com/platinasystems/go/coreutils/ln"
	"github.com/platinasystems/go/coreutils/log"
	"github.com/platinasystems/go/coreutils/ls"
	"github.com/platinasystems/go/coreutils/mkdir"
	"github.com/platinasystems/go/coreutils/ps"
	"github.com/platinasystems/go/coreutils/pwd"
	"github.com/platinasystems/go/coreutils/reboot"
	"github.com/platinasystems/go/coreutils/rm"
	"github.com/platinasystems/go/coreutils/sleep"
	"github.com/platinasystems/go/coreutils/stty"
	"github.com/platinasystems/go/coreutils/sync"
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
		reboot.New(),
		rm.New(),
		sleep.New(),
		stty.New(),
		sync.New(),
	}
}
