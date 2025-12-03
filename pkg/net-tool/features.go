// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tool

import (
	"github.com/platinasystems/goes/v2/pkg/net-tool/icmp"
	"github.com/platinasystems/goes/v2/pkg/net-tool/ifconfig"
	"github.com/platinasystems/goes/v2/pkg/net-tool/im"
	"github.com/platinasystems/goes/v2/pkg/net-tool/ndp"
	"github.com/platinasystems/goes/v2/pkg/net-tool/netcat"
	"github.com/platinasystems/goes/v2/pkg/net-tool/netstat"
	"github.com/platinasystems/goes/v2/pkg/net-tool/route"
	udp_echo "github.com/platinasystems/goes/v2/pkg/net-tool/udp-echo"
	"github.com/platinasystems/goes/v2/pkg/net-tool/wget"
	www_echo "github.com/platinasystems/goes/v2/pkg/net-tool/www-echo"
)

// Features mapped by name that emulate Unix net-tools.
var Features = map[string]any{
	"ifconfig": ifconfig.Ifconfig,
	"im":       im.InstantMessaging,
	"ndp":      ndp.NDP,
	"nc":       netcat.NetCat,
	"netcat":   netcat.NetCat,
	"netstat":  netstat.Netstat,
	"ping":     icmp.Ping,
	"route":    route.Features,
	"udp-echo": udp_echo.Features,
	"wget":     wget.Wget,
	"www-echo": www_echo.Features,
}
