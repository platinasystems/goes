// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"context"
	"errors"
	"net"
	"os"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

const MaxPacketSize = 1232

var ResetDeadline time.Time

var Resolver = &net.Resolver{
	PreferGo:     true,
	StrictErrors: true,
}

func NewUDP(ctx context.Context, svr string, port uint) (*net.UDPConn, error) {
	a := &net.UDPAddr{
		IP:   net.ParseIP(svr),
		Port: int(port),
	}
	if a.IP == nil {
		sl, err := Resolver.LookupIPAddr(ctx, svr)
		if err != nil {
			return nil, err
		}
		a.IP = sl[0].IP
		a.Zone = sl[0].Zone
	}
	nw := "udp"
	if a.IP.To4() != nil {
		nw = "udp4"
	} else if a.IP.To16() != nil {
		nw = "udp6"
	}
	return net.DialUDP(nw, nil, a)
}

// TenaciousAsk resends the buffered query every 1 sec until it receives a
// response or context is cancelled.
// If successful, it returns the raw binary response within the same,
// probably expanded buffer.
func TenaciousAsk(
	ctx context.Context, udp *net.UDPConn, buf []byte,
) ([]byte, error) {
	for {
		err := ctx.Err()
		if err != nil {
			return buf[:0], err
		}
		n, err := udp.Write(buf)
		if err != nil {
			return buf[:0], err
		} else if n != len(buf) {
			err = xerrors.Underrun("socket-write")
			return buf[:0], err
		}
		err = udp.SetReadDeadline(time.Now().Add(time.Second))
		if err != nil {
			return buf[:0], err
		}
		n, err = udp.Read(buf[:cap(buf)])
		udp.SetReadDeadline(ResetDeadline)
		if err == nil {
			return buf[:n], err
		} else if !errors.Is(err, os.ErrDeadlineExceeded) {
			return buf[:0], err
		}
	}
}

// TimeLimitedAsk is a `TenaciousAsk` with a deadlined context.
func TimeLimitedAsk(
	ctx context.Context, udp *net.UDPConn, buf []byte, dur time.Duration,
) ([]byte, error) {
	dl := time.Now().Add(dur)
	cctx, cancel := context.WithDeadline(ctx, dl)
	defer cancel()
	return TenaciousAsk(cctx, udp, buf)
}
