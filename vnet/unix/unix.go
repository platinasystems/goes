// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/vnet"
)

var packageIndex uint

type Main struct {
	vnet.Package

	v *vnet.Vnet

	verbosePackets bool
	verboseNetlink int

	netlink_main
	tuntapMain

	// For external (e.g. non tuntap) interfaces.
	siByIfIndex map[int]vnet.Si
}

func GetMain(v *vnet.Vnet) *Main { return v.GetPackage(packageIndex).(*Main) }

func Init(v *vnet.Vnet) {
	m := &Main{}
	m.v = v
	m.tuntapMain.Init(v)
	m.netlink_main.Init(m)
	packageIndex = v.AddPackage("tuntap", m)
}
