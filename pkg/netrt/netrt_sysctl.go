// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netrt

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"golang.org/x/sys/unix"
)

type list struct{ data []byte }

func NewList(ctx context.Context, family int) (Streamer, error) {
	data, err := xnet.SysctlGet(
		unix.CTL_NET,
		xnet.AF_ROUTE,
		0,
		int32(family),
		unix.NET_RT_DUMP,
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

func (l *list) Next(ctx context.Context) (NetRt, error) {
	for len(l.data) > xnet.SysctlMsgMin {
		rtm := PointerRtMsghdr(l.data)
		if len(l.data) < Sizeof(rtm) {
			break
		}
		n := xnet.SysctlAlign(int(rtm.Msglen))
		msg := l.data[:n]
		l.data = l.data[n:]
		if rtm.Version != unix.RTM_VERSION {
			continue
		}
		if rtm.Type != unix.RTM_GET {
			continue
		}
		// macOS seems to filter routes with both of these flags
		const gh = unix.RTF_GATEWAY | unix.RTF_HOST
		if (rtm.Flags & gh) == gh {
			continue
		}
		return newNetRt(msg), nil
	}
	return nil, nil
}
