// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	_ "embed"

	"errors"
)

//go:embed LICENSE
var license string

//go:embed PATENTS
var patents string

var ErrUnavailable = errors.New("unavailable")

var Features = map[string]any{
	"alarm":   AlarmDaemons,
	"command": Command,
	"cutoff":  Cutoff,
	"env":     Env,
	"errata":  Errata,
	"input":   Input,
	"log":     Log,
	"notice":  Notice,
	"output":  Output,
	"pty":     Pty,
	"show":    Shows,
	"standby": Standby,
	"start":   StartDaemon,
	"stop":    StopDaemons,
}

var Shows = map[string]any{
	"completion": ShowCompletion,
	"daemons":    ShowDaemons,
	"license":    license,
	"patents":    patents,
}
