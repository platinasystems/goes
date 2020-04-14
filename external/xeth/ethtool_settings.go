// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import "github.com/platinasystems/goes/external/xeth/internal"

type AutoNeg uint8
type Duplex uint8
type DevPort uint8

type DevEthtoolSettings Xid

func (xid Xid) RxEthtoolSettings(msg *internal.MsgEthtoolSettings) DevEthtoolSettings {
	l := LinkOf(xid)
	l.EthtoolSpeed(msg.Speed)
	l.EthtoolAutoNeg(AutoNeg(msg.Autoneg))
	l.EthtoolDuplex(Duplex(msg.Duplex))
	l.EthtoolDevPort(DevPort(msg.Port))
	return DevEthtoolSettings(xid)
}
