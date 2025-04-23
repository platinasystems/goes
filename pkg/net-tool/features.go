// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tool

import "github.com/platinasystems/goes/v2/pkg/net-tool/route"

// Features mapped by name that emulate Unix net-tools.
var Features = map[string]any{
	"ifconfig": Ifconfig,
	"ndp":      NDP,
	"netstat":  Netstat,
	"nslookup": Nslookup,
	"ping":     ICMPPing,
	"route":    route.Features,
	"udp-echo": UDPEcho,
	"udp-ping": UDPPing,
	"www-echo": WWWEcho,
	"www-ping": WWWPing,
}
