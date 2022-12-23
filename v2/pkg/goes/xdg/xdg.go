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
	"cache-home":   show.Text{xdg.Cache.CacheHome}.Func,
	"config-dirs":  show.Text{xdg.Cache.ConfigDirs}.Func,
	"config-home":  show.Text{xdg.Cache.ConfigHome}.Func,
	"data-dirs":    show.Text{xdg.Cache.DataDirs}.Func,
	"data-home":    show.Text{xdg.Cache.DataHome}.Func,
	"run-time-dir": show.Text{xdg.Cache.RunTimeDir}.Func,
	"state-home":   show.Text{xdg.Cache.StateHome}.Func,
}.Select
