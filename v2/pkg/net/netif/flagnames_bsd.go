// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build freebsd || openbsd || netbsd || darwin

package netif

var FlagNames = []string{
	"UP",
	"BROADCAST",
	"DEBUG",
	"LOOPBACK",
	"POINTOPOINT",
	"SMART/NOTRAILERS",
	"RUNNING",
	"NOARP",
	"PROMISC",
	"ALLMULTI",
	"OACTIVE",
	"SIMPLEX",
	"LINK0",
	"LINK1",
	"LINK2",
	"MULTICAST",
}
