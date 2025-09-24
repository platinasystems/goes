// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/goes/v2/pkg/bind"
	core_util "github.com/platinasystems/goes/v2/pkg/core-util"
	"github.com/platinasystems/goes/v2/pkg/goes"
	goes_util "github.com/platinasystems/goes/v2/pkg/goes-util"
	net_tool "github.com/platinasystems/goes/v2/pkg/net-tool"
	"github.com/platinasystems/goes/v2/pkg/vpn"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

var features = []map[string]any{
	goes_util.Features,
	core_util.Features,
	net_tool.Features,
	bind.Features,
	vpn.Features,
	xmain.Features,
	xprogram.Features,
}

func init() { goes.Install(features...) }

func main() { goes.Main() }
