// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package route

import (
	"fmt"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/netrt"
	"golang.org/x/sys/unix"
)

const gwFlags = `{blackhole, cloning, proto1, proto2, proto3, reject,
 xresolve, static}`

const gwMetrics = `{mtu, hopcount, expire, recvpipe, sendpipe, ssthresh,
rtt, rttvar}`

var gwFlagValues = map[string]uint{
	"blackhole": unix.RTF_BLACKHOLE,
	"cloning":   unix.RTF_CLONING,
	"proto1":    unix.RTF_PROTO1,
	"proto2":    unix.RTF_PROTO2,
	"proto3":    unix.RTF_PROTO3,
	"reject":    unix.RTF_REJECT,
	"xresolve":  unix.RTF_XRESOLVE,
	"static":    unix.RTF_STATIC,
}

func rtmdel(rtm *netrt.RtMsghdr2) {
	integer.Set(&rtm.Flags, unix.RTF_PINNED)
}

func rtmmetrics(rtm *netrt.RtMsghdr2, metrics []string) {
	for _, metric := range metrics {
		eq := strings.Index(metric, "=")
		if eq < 1 {
			continue
		}
		name := metric[:eq]
		var val uint
		if _, err := fmt.Sscan(metric[eq+1:], &val); err != nil {
			continue
		}
		switch name {
		case "mtu":
			if val != 1500 {
				integer.Assign(&rtm.Rmx.Mtu, val)
				rtm.Inits |= unix.RTV_MTU
			}
		case "hopcount":
			integer.Assign(&rtm.Rmx.Hopcount, val)
			rtm.Inits |= unix.RTV_HOPCOUNT
		case "expire":
			elapse := time.Second * time.Duration(int(val))
			integer.Assign(&rtm.Rmx.Expire,
				time.Now().Add(elapse).Unix())
			rtm.Inits |= unix.RTV_EXPIRE
		case "recvpipe":
			integer.Assign(&rtm.Rmx.Recvpipe, val)
			rtm.Inits |= unix.RTV_RPIPE
		case "sendpipe":
			integer.Assign(&rtm.Rmx.Sendpipe, val)
			rtm.Inits |= unix.RTV_SPIPE
		case "ssthresh":
			integer.Assign(&rtm.Rmx.Ssthresh, val)
			rtm.Inits |= unix.RTV_SSTHRESH
		case "rtt":
			integer.Assign(&rtm.Rmx.Rtt, val)
			rtm.Inits |= unix.RTV_RTT
		case "rttvar":
			integer.Assign(&rtm.Rmx.Rttvar, val)
			rtm.Inits |= unix.RTV_RTTVAR
		}
	}
}
