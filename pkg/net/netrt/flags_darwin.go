// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import (
	"context"
	"syscall"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
)

func setRtmFlags(ctx context.Context, cmd uint8, rtm *syscall.RtMsghdr) {
	switch cmd {
	case syscall.RTM_ADD, syscall.RTM_CHANGE:
		rtm.Flags = syscall.RTF_UP
	case syscall.RTM_DELETE:
		rtm.Flags |= syscall.RTF_PINNED
	}
	for _, x := range []struct {
		name string
		flag int32
	}{
		{"static", syscall.RTF_STATIC},
		{"reject", syscall.RTF_REJECT},
		{"blackhole", syscall.RTF_BLACKHOLE},
		{"proto1", syscall.RTF_PROTO1},
		{"proto2", syscall.RTF_PROTO2},
		{"proxy", syscall.RTF_PROXY},
		{"xresolve", syscall.RTF_XRESOLVE},
	} {
		if ctxparm.SearchFlagsIn[bool](ctx, x.name) {
			rtm.Flags |= x.flag
		}
	}
	if !ctxparm.SearchFlagsIn[bool](ctx, "interface") {
		rtm.Flags |= syscall.RTF_GATEWAY
	}
	if mtu := ctxparm.SearchFlagsIn[int](ctx, "mtu"); mtu != 1500 {
		rtm.Rmx.Mtu = uint32(mtu)
		rtm.Inits |= syscall.RTV_MTU
	}
	if secs := ctxparm.SearchFlagsIn[int](ctx, "expire"); secs != 0 {
		elapse := time.Second * time.Duration(secs)
		expire := time.Now().Add(elapse).Unix()
		rtm.Rmx.Expire = int32(expire)
		rtm.Inits |= syscall.RTV_EXPIRE
	}
	for _, x := range []struct {
		name string
		rtv  uint32
		rmx  *uint32
	}{
		{"hopcount", syscall.RTV_HOPCOUNT, &rtm.Rmx.Hopcount},
		{"recvpipe", syscall.RTV_RPIPE, &rtm.Rmx.Recvpipe},
		{"sendpipe", syscall.RTV_SPIPE, &rtm.Rmx.Sendpipe},
		{"ssthresh", syscall.RTV_SSTHRESH, &rtm.Rmx.Ssthresh},
		{"rtt", syscall.RTV_RTT, &rtm.Rmx.Rtt},
		{"rttvar", syscall.RTV_RTTVAR, &rtm.Rmx.Rttvar},
	} {
		if v := ctxparm.SearchFlagsIn[int](ctx, x.name); v != 0 {
			*(x.rmx) = uint32(v)
			rtm.Inits |= x.rtv
		}
	}
}
