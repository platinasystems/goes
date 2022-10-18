// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdg

import (
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/goes/show"
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
)

var Show = selection.Map{
	"cache-home":   show.Func(xdg.CacheHome),
	"config-dirs":  show.Func(xdg.ConfigDirs),
	"config-home":  show.Func(xdg.ConfigHome),
	"data-dirs":    show.Func(xdg.DataDirs),
	"data-home":    show.Func(xdg.DataHome),
	"run-time-dir": show.Func(xdg.RunTimeDir),
	"state-home":   show.Func(xdg.StateHome),
}.Select
