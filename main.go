// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/goes/v2/pkg/bind"
	"github.com/platinasystems/goes/v2/pkg/cert"
	core_util "github.com/platinasystems/goes/v2/pkg/core-util"
	"github.com/platinasystems/goes/v2/pkg/goes"
	goes_util "github.com/platinasystems/goes/v2/pkg/goes-util"
	net_tool "github.com/platinasystems/goes/v2/pkg/net-tool"
	"github.com/platinasystems/goes/v2/pkg/sig"
	"github.com/platinasystems/goes/v2/pkg/vpn"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

func init() {
	goes.Install(
		xmain.Features,
		xprogram.Features,
		goes_util.Features,
		core_util.Features,
		net_tool.Features,
		bind.Features,
		sig.Features,
		cert.Features,
		vpn.Features,
	)
}

func main() { goes.Main() }
