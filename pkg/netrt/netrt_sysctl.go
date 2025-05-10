// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netrt

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xos/sysctl"
	"golang.org/x/sys/unix"
)

type sysctlRsp struct {
	data []byte
}

func Neighbors(ctx context.Context, family int) (NextNdCloser, error) {
	mib := []int32{
		sysctl.CTL_NET,
		xnet.AF_ROUTE,
		0,
		int32(family),
		unix.NET_RT_FLAGS,
		unix.RTF_LLINFO,
	}
	data, err := sysctl.Get(mib...)
	if err != nil {
		return nil, err
	}
	return &sysctlRsp{data}, nil
}

func Routes(ctx context.Context, family int) (NextRtCloser, error) {
	mib := []int32{
		sysctl.CTL_NET,
		xnet.AF_ROUTE,
		0,
		int32(family),
		unix.NET_RT_DUMP,
		0,
	}
	data, err := sysctl.Get(mib...)
	if err != nil {
		return nil, err
	}
	return &sysctlRsp{data}, nil
}

func (rsp *sysctlRsp) Close() error {
	rsp.data = []byte{}[:0]
	return nil
}

func (rsp *sysctlRsp) NextNd(ctx context.Context) (Nd, error) {
	for len(rsp.data) > sysctl.Min {
		rtm, body, data := Extract[RtMsghdr2](rsp.data)
		if rtm == nil {
			break
		}
		rsp.data = data
		if rtm.Version != unix.RTM_VERSION {
			continue
		}
		if rtm.Type != unix.RTM_GET {
			continue
		}
		if (rtm.Flags & unix.RTF_HOST) == 0 {
			continue
		}
		return newNetRt(rtm, body), nil
	}
	return nil, nil
}

func (rsp *sysctlRsp) NextRt(ctx context.Context) (Rt, error) {
	const gh = unix.RTF_GATEWAY | unix.RTF_HOST
	for len(rsp.data) > sysctl.Min {
		rtm, body, data := Extract[RtMsghdr2](rsp.data)
		if rtm == nil {
			break
		}
		rsp.data = data
		if rtm.Version != unix.RTM_VERSION {
			continue
		}
		if rtm.Type != unix.RTM_GET {
			continue
		}
		if (rtm.Flags & gh) == gh {
			continue
		}
		return newNetRt(rtm, body), nil
	}
	return nil, nil
}
