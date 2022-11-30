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
	"cache-home":   show.New(xdg.Cache.CacheHome),
	"config-dirs":  show.New(xdg.Cache.ConfigDirs),
	"config-home":  show.New(xdg.Cache.ConfigHome),
	"data-dirs":    show.New(xdg.Cache.DataDirs),
	"data-home":    show.New(xdg.Cache.DataHome),
	"run-time-dir": show.New(xdg.Cache.RunTimeDir),
	"state-home":   show.New(xdg.Cache.StateHome),
}.Select
