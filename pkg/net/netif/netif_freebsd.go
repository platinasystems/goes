// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"syscall"
	"time"
)

const SIOCIFCREATE = syscall.SIOCIFCREATE2

type IfmaMsgHdr = syscall.IfmaMsghdr

type Msgs interface {
	IfMsgHdr | IfaMsgHdr | IfmaMsgHdr
}

func (nif *NetIf) ExtraIfData(ifdata *syscall.IfData) {
	if ifdata.Metric != 0 {
		nif.Extra["metric"] = ifdata.Metric
	}
	if ifdata.Baudrate != 0 {
		nif.Extra["baudrate"] = ifdata.Baudrate
	}
	if ifdata.Imcasts != 0 {
		nif.Extra["imcasts"] = ifdata.Imcasts
	}
	if ifdata.Omcasts != 0 {
		nif.Extra["omcasts"] = ifdata.Omcasts
	}
	if ifdata.Noproto != 0 {
		nif.Extra["noproto"] = ifdata.Noproto
	}
	if ifdata.Lastchange.Sec != 0 {
		nif.Extra["lastchange"] = time.Unix(
			int64(ifdata.Lastchange.Sec),
			int64(ifdata.Lastchange.Usec)*1000)
	}
	if ifdata.Hwassist != 0 {
		nif.Extra["hwassist"] = ifdata.Hwassist
	}
}
