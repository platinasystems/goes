// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"errors"
	"golang/buildid"

	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

var ErrUnavailable = errors.New("unavailable")

var Features = map[string]any{
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
			"id": func() (any, error) {
				return buildid.ReadFile(xprogram.Path())
			},
			"info": func() (any, error) {
				var err error
				bi := xprogram.BuildInfo()
				if bi == nil {
					err = ErrUnavailable
				}
				return bi, err
			},
		},
		"daemons": ShowDaemons,
		"main": map[string]any{
			"name": func() (any, error) {
				return xprogram.MainName(), nil
			},
			"reference": func() (any, error) {
				m := xprogram.MainModule()
				if m == nil {
					return "", ErrUnavailable
				}
				return m.Path + "@" + m.Version, nil
			},
			"version": func() (any, error) {
				m := xprogram.MainModule()
				if m == nil {
					return "", ErrUnavailable
				}
				return m.Version, nil
			},
		},
	},
	"standby": Standby,
	"start":   StartDaemon,
	"stop":    StopDaemons,
}
