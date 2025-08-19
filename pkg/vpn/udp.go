// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"net"

	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

// Open socket, make channels and start packet read/write routines.
// If not error, close the returned write channel when done.
func startUDP(ctx context.Context, port uint16) (
	<-chan *xnet.Msg, chan<- *xnet.Msg, error,
) {
	fromC := make(chan *xnet.Msg, FromVpnCap)
	toC := make(chan *xnet.Msg, ToVpnCap)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: int(port),
	})
	if err != nil {
		return nil, nil, err
	}
	wg.Go(func() {
		const kind = "receive service"
		xlog.Trace.Println("start", port, kind)
		err := mp.RecvBatchService(ctx, fromC, conn)
		if err != nil {
			xlog.Errata.Println("quit", port, kind, err)
		} else {
			xlog.Trace.Println("stopped", port, kind)
		}
	})
	wg.Go(func() {
		const kind = "send service"
		xlog.Trace.Println("start", port, kind)
		err := mp.SendBatchService(conn, toC)
		if err != nil {
			xlog.Errata.Println("quit", port, kind, err)
		} else {
			xlog.Trace.Println("stopped", port, kind)
		}
		conn.Close()
	})
	return fromC, toC, err
}
