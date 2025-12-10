// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package core-util contain [goes.Features] that mimic Unix coreutils.
package core_util

import (
	"github.com/platinasystems/goes/v2/pkg/core-util/cat"
	"github.com/platinasystems/goes/v2/pkg/core-util/chmod"
	"github.com/platinasystems/goes/v2/pkg/core-util/cp"
	"github.com/platinasystems/goes/v2/pkg/core-util/echo"
	"github.com/platinasystems/goes/v2/pkg/core-util/env"
	"github.com/platinasystems/goes/v2/pkg/core-util/hostname"
	"github.com/platinasystems/goes/v2/pkg/core-util/id"
	"github.com/platinasystems/goes/v2/pkg/core-util/kill"
	"github.com/platinasystems/goes/v2/pkg/core-util/ln"
	"github.com/platinasystems/goes/v2/pkg/core-util/ls"
	"github.com/platinasystems/goes/v2/pkg/core-util/mkdir"
	"github.com/platinasystems/goes/v2/pkg/core-util/mknod"
	"github.com/platinasystems/goes/v2/pkg/core-util/pwd"
	"github.com/platinasystems/goes/v2/pkg/core-util/rm"
	"github.com/platinasystems/goes/v2/pkg/core-util/sleep"
	"github.com/platinasystems/goes/v2/pkg/core-util/stty"
	"github.com/platinasystems/goes/v2/pkg/core-util/sync"
	"github.com/platinasystems/goes/v2/pkg/core-util/uptime"
)

var Features = map[string]any{
	"cat":      cat.Cat,
	"chmod":    chmod.Chmod,
	"cp":       cp.Cp,
	"echo":     echo.Echo,
	"env":      env.Env,
	"hostname": hostname.Hostname,
	"id":       id.Id,
	"kill":     kill.Kill,
	"ln":       ln.Ln,
	"ls":       ls.Ls,
	"mkdir":    mkdir.Mkdir,
	"mknod":    mknod.Mknod,
	"pwd":      pwd.Pwd,
	"rm":       rm.Rm,
	"sleep":    sleep.Sleep,
	"stty":     stty.Stty,
	"sync":     sync.Sync,
	"uptime":   uptime.Uptime,
}
