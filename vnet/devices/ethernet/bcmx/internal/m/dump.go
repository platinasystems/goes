// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package m

// Configurable port settings.
//
// Some assumptions made here:
// interface doesn't seem to alter phy settings (e.g. KR4 vs CR4
// encap for us is BCM_PORT_ENCAP_IEEE
type portFlags uint32

const (
	portEnable portFlags = 1 << iota
	portAutoneg
	portSpeed
	portLoopback
)

type portSettings struct {
	enable          bool
	autoneg         bool
	speedBitsPerSec float64
	loopback_type   PortLoopbackType
}

func portSetDefault(p Porter) (mask portFlags, ps portSettings) {

	mask |= portEnable
	ps.enable = true

	mask |= portAutoneg
	ps.autoneg = true

	mask |= portSpeed
	ps.speedBitsPerSec = 100e9

	mask |= portLoopback
	ps.loopback_type = PortLoopbackPhyLocal

	return
}

func portPhySettings(phy Phyer, port Porter, mask portFlags, ps *portSettings) {

	for i := 0; i < 32; i++ {
		switch mask & (1 << uint32(i)) {
		case portAutoneg:
			phy.SetAutoneg(port, true)
		case portSpeed:
			phy.SetSpeed(port, ps.speedBitsPerSec, false)
		case portLoopback:
			phy.SetLoopback(port, ps.loopback_type)
		}
	}
}

const (
	PORT_PHY_LOCAL_LOOPBACK uint32 = iota
	PORT_PHY_REMOTE_LOOPBACK
	PORT_MAC_LOOPBACK
	PORT_NO_LOOPBACK
)
