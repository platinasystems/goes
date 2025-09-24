// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package goes-vpn provides a standalone [vpn].
package main

import (
	"github.com/platinasystems/goes/v2/pkg/goes"
	goes_util "github.com/platinasystems/goes/v2/pkg/goes-util"
	"github.com/platinasystems/goes/v2/pkg/vpn"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

var features = []map[string]any{
	vpn.MainFeatures,
	map[string]any{
		"new":  vpn.NewFeatures,
		"show": vpn.ShowFeatures,
	},
	goes_util.Features,
	xmain.Features,
	xprogram.Features,
}

func init() { goes.Install(features...) }

func main() { goes.Main() }
