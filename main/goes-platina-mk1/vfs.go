// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import "github.com/platinasystems/go/internal/sriovs"

func vlan_for_port(port, subport sriovs.Vf) (vf sriovs.Vf) {
	// physical port number for data ports are numbered starting at 1.
	// (phys 0 is cpu port...)
	phys := sriovs.Vf(1)

	// 4 sub-ports per port; mk1 ports are even/odd swapped.
	phys += 4 * (port ^ 1)

	phys += subport

	// Vlan is 1 plus physical port number.
	return sriovs.Vf(1 + phys)
}

// The vfs table is 0 based and is adjusted to 1 based beta and production
// units with VfName
var vfs = make_vfs()

func make_vfs() [][]sriovs.Vf {
	// pf0 = fe1 pipes 0 & 1; only 63 vfs supported so last sub port is not accessible.
	// pf1 = fe1 pipes 2 & 3; only 63 vfs supported so last sub port is not accessible.
	var pfs [2][63]sriovs.Vf
	for port := sriovs.Vf(0); port < 32; port++ {
		for subport := sriovs.Vf(0); subport < 4; subport++ {
			vf := port<<sriovs.PortShift | subport<<sriovs.SubPortShift | vlan_for_port(port, subport)
			pf := port / 16
			i := 4*(port%16) + subport
			if i < sriovs.Vf(len(pfs[pf])) {
				pfs[pf][i] = vf
			}
		}
	}
	return [][]sriovs.Vf{pfs[0][:], pfs[1][:]}
}
