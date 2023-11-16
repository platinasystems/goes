// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || freebsd || netbsd

package netioctl

import (
	"fmt"
	"syscall"
)

const (
	SIOCGIFCAP = syscall.SIOCGIFCAP
	SIOCSIFCAP = syscall.SIOCSIFCAP
)

const (
	IFCAPsize = 4
	IFCAPbits = IFCAPsize * 8
)

type IFCAP uint32

const (
	IFCAP_RXCSUM IFCAP = 1 << iota
	IFCAP_TXCSUM
	IFCAP_VLAN_MTU
	IFCAP_VLAN_HWTAGGING
	IFCAP_JUMBO_MTU
	IFCAP_TSO4
	IFCAP_TSO6
	IFCAP_LRO
	IFCAP_AV
	IFCAP_TXSTATUS
	IFCAP_SKYWALK
	IFCAP_HW_TIMESTAMP
	IFCAP_SW_TIMESTAMP
	IFCAP_CSUM_PARTIAL
	IFCAP_CSUM_ZERO_INVERT
)

type IfCaps struct{ Req, Cur IFCAP }

func (ifcap IFCAP) Format(w fmt.State, verb rune) {
	if verb == 'x' {
		fmt.Fprintf(w, "%#04x", uint(ifcap))
		return
	}
	if verb != 's' && verb != 'v' {
		return
	}
	var comma string
	for bit := 0; bit < IFCAPbits; bit++ {
		if flag := IFCAP(1 << bit); (ifcap & flag) != 0 {
			fmt.Fprint(w, comma, flag.String())
			comma = ","
		}
	}
}
