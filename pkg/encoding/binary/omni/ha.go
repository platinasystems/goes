// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package omni

import "net"

type HardwareAddr [6]byte

func (ha *HardwareAddr) Put(v net.HardwareAddr) { copy(ha[:], v[:]) }

func (ha *HardwareAddr) Value() net.HardwareAddr {
	return net.HardwareAddr(ha[:])
}
