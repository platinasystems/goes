// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package goes-vpn provides a standalone [vpn].
package main

import (
	"github.com/platinasystems/goes/v2/pkg/bind/dig"
	"github.com/platinasystems/goes/v2/pkg/bind/host"
	"github.com/platinasystems/goes/v2/pkg/bind/nslookup"
	"github.com/platinasystems/goes/v2/pkg/cert"
	"github.com/platinasystems/goes/v2/pkg/goes"
	goes_util "github.com/platinasystems/goes/v2/pkg/goes-util"
	"github.com/platinasystems/goes/v2/pkg/net-tool/icmp"
	"github.com/platinasystems/goes/v2/pkg/net-tool/im"
	"github.com/platinasystems/goes/v2/pkg/net-tool/netcat"
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
		sig.Features,
		cert.Features,
		vpn.MainFeatures,
		map[string]any{
			"dig":      dig.Dig,
			"host":     host.Host,
			"im":       im.InstantMessaging,
			"nc":       netcat.NetCat,
			"nslookup": nslookup.Nslookup,
			"netcat":   netcat.NetCat,
			"ping":     icmp.Ping,
			"show":     vpn.ShowFeatures,
		},
	)
}

func main() { goes.Main() }
