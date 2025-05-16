// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"errors"

	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

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
	"show": map[string]any{
		"build": map[string]any{
			"id":      xprogram.BuildId,
			"info":    xprogram.BuildInfoString,
			"setting": xprogram.BuildSettings,
		},
		"daemons": ShowDaemons,
		"main": map[string]any{
			"name":      xprogram.MainName,
			"reference": xprogram.MainReference,
			"version":   xprogram.MainVersion,
		},
	},
	"standby": Standby,
	"start":   StartDaemon,
	"stop":    StopDaemons,
}
