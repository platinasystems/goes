// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netrt

import (
	"context"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/net/sysctl"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

type list struct{ data []byte }

func NewList(ctx context.Context) (Streamer, error) {
	var family int32
	if ctxparm.SearchFlagsIn[bool](ctx, "4") {
		family = af.INET
	} else if ctxparm.SearchFlagsIn[bool](ctx, "6") {
		family = af.INET6
	} else {
		family = af.UNSPEC
	}
	data, err := sysctl.Get(
		syscall.CTL_NET,
		af.ROUTE,
		0,
		family,
		syscall.NET_RT_DUMP,
		0,
		// FIXME fib?
	)
	if err != nil {
		return nil, err
	}
	return &list{data}, nil
}

func (l *list) Close() error {
	l.data = l.data[:0]
	return nil
}

func (l *list) Next() (NetRt, error) {
	for len(l.data) > sysctl.MsgMin {
		rtm := PointerRtMsghdr(l.data)
		if len(l.data) < Sizeof(rtm) {
			break
		}
		n := sysctl.Align(int(rtm.Msglen))
		msg := l.data[:n]
		l.data = l.data[n:]
		if rtm.Version != syscall.RTM_VERSION {
			continue
		}
		if rtm.Type != syscall.RTM_GET {
			continue
		}
		// macOS seems to filter routes with both of these flags
		const gh = syscall.RTF_GATEWAY | syscall.RTF_HOST
		if (rtm.Flags & gh) == gh {
			continue
		}
		return newNetRt(msg), nil
	}
	return nil, nil
}
