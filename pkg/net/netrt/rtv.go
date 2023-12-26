// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux && !freebsd

package netrt

import (
	"flag"
	"time"

	"github.com/platinasystems/goes/v2/pkg/flag/flagset"
	"github.com/platinasystems/goes/v2/pkg/integer"
)

const (
	RTV_MTU = 1 << iota
	RTV_HOPCOUNT
	RTV_EXPIRE
	RTV_RPIPE
	RTV_SPIPE
	RTV_SSTHRESH
	RTV_RTT
	RTV_RTTVAR
)

func AddMetricFlags(opts *flag.FlagSet) {
	opts.Uint("mtu", 1500, "FIXME")
	opts.Uint("hopcount", 0, "FIXME")
	opts.Int("expire", 0, "Seconds from now.")
	opts.Uint("recvpipe", 0, "FIXME")
	opts.Uint("sendpipe", 0, "FIXME")
	opts.Uint("ssthresh", 0, "FIXME")
	opts.Uint("rtt", 0, "FIXME")
	opts.Uint("rttvar", 0, "FIXME")
}

func SetMetrics(rtm *RtMsghdr, opts *flag.FlagSet) {
	if u := flagset.Search[uint](opts, "mtu"); u != 1500 {
		integer.Assign(&rtm.Rmx.Mtu, u)
		rtm.Inits |= RTV_MTU
	}
	if u := flagset.Search[uint](opts, "hopcount"); u != 0 {
		integer.Assign(&rtm.Rmx.Hopcount, u)
		rtm.Inits |= RTV_HOPCOUNT
	}
	if secs := flagset.Search[int](opts, "expire"); secs != 0 {
		elapse := time.Second * time.Duration(secs)
		integer.Assign(&rtm.Rmx.Expire, time.Now().Add(elapse).Unix())
		rtm.Inits |= RTV_EXPIRE
	}
	if u := flagset.Search[uint](opts, "recvpipe"); u != 0 {
		integer.Assign(&rtm.Rmx.Recvpipe, u)
		rtm.Inits |= RTV_RPIPE
	}
	if u := flagset.Search[uint](opts, "sendpipe"); u != 0 {
		integer.Assign(&rtm.Rmx.Sendpipe, u)
		rtm.Inits |= RTV_SPIPE
	}
	if u := flagset.Search[uint](opts, "ssthresh"); u != 0 {
		integer.Assign(&rtm.Rmx.Ssthresh, u)
		rtm.Inits |= RTV_SSTHRESH
	}
	if u := flagset.Search[uint](opts, "rtt"); u != 0 {
		integer.Assign(&rtm.Rmx.Rtt, u)
		rtm.Inits |= RTV_RTT
	}
	if u := flagset.Search[uint](opts, "rttvar"); u != 0 {
		integer.Assign(&rtm.Rmx.Rttvar, u)
		rtm.Inits |= RTV_RTTVAR
	}
}
