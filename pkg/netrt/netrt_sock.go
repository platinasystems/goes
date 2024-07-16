// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netrt

import (
	"context"
	"errors"
	"os"
	"sync/atomic"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"golang.org/x/sys/unix"
)

type appendAddrFunc func(context.Context, []byte) ([]byte, error)

var seq atomic.Int32

func Flush(ctx context.Context) error {
	return FIXME
}

func Monitor(ctx context.Context) (Streamer, error) {
	return nil, FIXME
}

func NewRtMsg() []byte {
	pgsz := os.Getpagesize()
	msg := make([]byte, pgsz, pgsz)
	rtm := PointerRtMsghdr(msg)
	msg = msg[:Sizeof(rtm)]
	// rtm.Type = rtmt
	rtm.Version = unix.RTM_VERSION
	rtm.Seq = seq.Add(1)
	return msg
}

func Request(msg []byte, fib int) (NetRt, error) {
	rtm := PointerRtMsghdr(msg)
	rtm.Msglen = uint16(len(msg))
	sock, err := xnet.OpenRoute()
	if err != nil {
		return nil, xerrors.Mark(err)
	}
	defer xnet.Close(sock)
	if fib >= 0 {
		if err = SetFib(sock, fib); err != nil {
			return nil, err
		}
	}
	if _, err = xnet.Write(sock, msg); err != nil {
		switch {
		case errors.Is(err, unix.ESRCH):
			return nil, ErrSRCH
		case errors.Is(err, unix.EBUSY):
			return nil, ErrBUSY
		case errors.Is(err, unix.ENOBUFS):
			return nil, ErrNOBUFS
		case errors.Is(err, unix.EADDRINUSE):
			return nil, ErrADDRINUSE
		case errors.Is(err, unix.EEXIST):
			return nil, ErrEXIST
		default:
			return nil, xerrors.Markf("%w\n%#v", err, rtm)
		}
	} else if rtm.Type != unix.RTM_GET {
		return nil, nil
	}
	msg = msg[:cap(msg)]
	n, err := xnet.Read(sock, msg)
	if err != nil {
		return nil, xerrors.Mark(err)
	}
	return newNetRt(msg[:n]), nil
}
