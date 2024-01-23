// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"flag"
	"time"

	"github.com/platinasystems/goes/v2/pkg/flag/flagset"
	"github.com/platinasystems/goes/v2/pkg/integer"
	"golang.org/x/sys/unix"
)

var routeGatewayFlags = []routeGatewayFlag{
	{unix.RTF_BLACKHOLE, "blackhole", false, "Silently discard packets."},
	{unix.RTF_CLONING, "cloning", false, "Generates a new route."},
	{unix.RTF_PROTO1, "proto1", false, "FIXME"},
	{unix.RTF_PROTO2, "proto2", false, "FIXME"},
	{unix.RTF_PROTO3, "proto3", false, "FIXME"},
	{unix.RTF_REJECT, "reject", false, "Emit ICMP unreachable."},
	{unix.RTF_XRESOLVE, "xresolve", false, "Emit mesg on use."},
	{unix.RTF_STATIC, "static", true, "Manually added route."},
}

var routeGatewayMetrics = []routeGatewayMetric{
	{"mtu", 1500, "FIXME"},
	{"hopcount", 0, "FIXME"},
	{"expire", 0, "Seconds from now."},
	{"recvpipe", 0, "FIXME"},
	{"sendpipe", 0, "FIXME"},
	{"ssthresh", 0, "FIXME"},
	{"rtt", 0, "FIXME"},
	{"rttvar", 0, "FIXME"},
}

func routeSetDelFlag(rtm *unix.RtMsghdr) {
	integer.Set(&rtm.Flags, unix.RTF_PINNED)
}

func routeSetMetrics(rtm *unix.RtMsghdr, opts *flag.FlagSet) {
	if u := flagset.Search[uint](opts, "mtu"); u != 1500 {
		integer.Assign(&rtm.Rmx.Mtu, u)
		rtm.Inits |= unix.RTV_MTU
	}
	if u := flagset.Search[uint](opts, "hopcount"); u != 0 {
		integer.Assign(&rtm.Rmx.Hopcount, u)
		rtm.Inits |= unix.RTV_HOPCOUNT
	}
	if secs := flagset.Search[int](opts, "expire"); secs != 0 {
		elapse := time.Second * time.Duration(secs)
		integer.Assign(&rtm.Rmx.Expire, time.Now().Add(elapse).Unix())
		rtm.Inits |= unix.RTV_EXPIRE
	}
	if u := flagset.Search[uint](opts, "recvpipe"); u != 0 {
		integer.Assign(&rtm.Rmx.Recvpipe, u)
		rtm.Inits |= unix.RTV_RPIPE
	}
	if u := flagset.Search[uint](opts, "sendpipe"); u != 0 {
		integer.Assign(&rtm.Rmx.Sendpipe, u)
		rtm.Inits |= unix.RTV_SPIPE
	}
	if u := flagset.Search[uint](opts, "ssthresh"); u != 0 {
		integer.Assign(&rtm.Rmx.Ssthresh, u)
		rtm.Inits |= unix.RTV_SSTHRESH
	}
	if u := flagset.Search[uint](opts, "rtt"); u != 0 {
		integer.Assign(&rtm.Rmx.Rtt, u)
		rtm.Inits |= unix.RTV_RTT
	}
	if u := flagset.Search[uint](opts, "rttvar"); u != 0 {
		integer.Assign(&rtm.Rmx.Rttvar, u)
		rtm.Inits |= unix.RTV_RTTVAR
	}
}
