// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

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
				return BuildId()
			},
			"info": func() (any, error) {
				return BuildInfo()
			},
		},
		"main": map[string]any{
			"reference": func() (any, error) {
				return MainReference()
			},
			"version": func() (any, error) {
				return MainVersion()
			},
		},
	},
	"standby": Standby,
	"start":   Start,
}
