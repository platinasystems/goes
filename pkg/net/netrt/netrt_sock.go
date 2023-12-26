// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netrt

import (
	"context"
	"errors"
	"sync/atomic"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/os/page"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
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
	msg := page.New()
	rtm := PointerRtMsghdr(msg)
	msg = msg[:Sizeof(rtm)]
	// rtm.Type = rtmt
	rtm.Version = syscall.RTM_VERSION
	rtm.Seq = seq.Add(1)
	return msg
}

func Request(msg []byte, fib int) (NetRt, error) {
	rtm := PointerRtMsghdr(msg)
	rtm.Msglen = uint16(len(msg))
	sock, err := af.OpenRoute()
	if err != nil {
		return nil, egress.Mark(err)
	}
	defer af.Close(sock)
	if fib >= 0 {
		err = FIXME
		// FIXME darwin doesn't have SO_SETFIB so contrain like this
		// err = sock.SetFib(fib)
		// err = os.NewSyscallError("SO_SETFIB", syscall.
		// 	SetsockoptInt(int(sock), syscall.SOL_SOCKET,
		// 		syscall.SO_SETFIB, fib))
		if err != nil {
			return nil, err
		}
	}
	if _, err = af.Write(sock, msg); err != nil {
		switch {
		case errors.Is(err, syscall.ESRCH):
			return nil, ErrSRCH
		case errors.Is(err, syscall.EBUSY):
			return nil, ErrBUSY
		case errors.Is(err, syscall.ENOBUFS):
			return nil, ErrNOBUFS
		case errors.Is(err, syscall.EADDRINUSE):
			return nil, ErrADDRINUSE
		case errors.Is(err, syscall.EEXIST):
			return nil, ErrEXIST
		default:
			return nil, egress.Markf("%w\n%#v", err, rtm)
		}
	} else if rtm.Type != syscall.RTM_GET {
		return nil, nil
	}
	msg = msg[:cap(msg)]
	n, err := af.Read(sock, msg)
	if err != nil {
		return nil, egress.Mark(err)
	}
	return newNetRt(msg[:n]), nil
}

/*FIXME
func rtrmx[T int32 | uint32](rmx *T, inits *uint32, v uint, rtv uint32) {
	if v > 0 {
		*(rmx) = T(v)
		*(inits) |= rtv
	}
}
*/
