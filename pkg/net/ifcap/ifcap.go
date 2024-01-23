// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || freebsd || netbsd

package ifcap

import "golang.org/x/sys/unix"

const (
	SIOCGIFCAP = unix.SIOCGIFCAP
	SIOCSIFCAP = unix.SIOCSIFCAP
)

var Strings = []string{
	"rxcsum",
	"txcsum",
	"vlan_mtu",
	"vlan_hwtagging",
	"jumbo_mtu",
	"tso4",
	"tso6",
	"lro",
	"av",
	"txstatus",
	"skywalk",
	"hw_timestamp",
	"sw_timestamp",
	"csum_partial",
	"csum_zero_invert",
}
