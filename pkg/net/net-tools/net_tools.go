// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

var Daemons = map[string]any{
	"udp-echo":  UDPEcho,
	"www-echo":  WWWEcho,
	"tuntapper": TunTapper,
}

var Root = map[string]any{
	"icmp-ping": ICMPPing,
	"ifconfig":  Ifconfig,
	"netstat":   Netstat,
	"route":     Route,
	"udp-ping":  UDPPing,
	"www-ping":  WWWPing,
}
