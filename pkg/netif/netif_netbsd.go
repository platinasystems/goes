// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import "golang.org/x/sys/unix"

const SIOCIFCREATE = unix.SIOCIFCREATE

type Msgs interface {
	IfMsgHdr | IfaMsgHdr
}

func (nif *NetIf) ExtraIfData(ifdata *unix.IfData) {
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
}
