// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package neutral

import "net"

type HardwareAddr [6]byte

func NewHardwareAddr(b []byte) (*HardwareAddr, []byte) {
	return (*HardwareAddr)(b), b[6:]
}

func (p *HardwareAddr) Put(v net.HardwareAddr) { copy(p[:], v[:]) }

func (p *HardwareAddr) Value() net.HardwareAddr {
	return net.HardwareAddr(p[:])
}
