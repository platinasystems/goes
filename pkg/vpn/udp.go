// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"net"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/xcontext"
	"github.com/platinasystems/goes/v2/pkg/xlog"
)

func pktRx(ctx context.Context, conn *net.UDPConn, ch chan<- *box.Box) {
	la := conn.LocalAddr()
	xlog.Info.Printf("start %v rx routine", la)
	defer xlog.Info.Printf("stopped %v rx routine", la)
	defer close(ch)
	bx := box.New()
	for !xcontext.IsDone(ctx) {
		if err := bx.Rx(conn); err == nil {
			xlog.Trace.Printf("rx[%d] %v <- %v",
				bx.Len(), la, bx.AddrPort)
			if xcontext.Queue(ctx, ch, bx) {
				bx = box.New()
			} else {
				break
			}
		} else {
			xlog.Errata.Print(err)
			break
		}
	}
	bx.Return()
}

func pktTx(ctx context.Context, conn *net.UDPConn, ch <-chan *box.Box) {
	la := conn.LocalAddr()
	xlog.Info.Printf("start %v tx routine", la)
	defer xlog.Info.Printf("stopped %v tx routine", la)
	xcontext.Range(ctx, ch, func(bx *box.Box) (ok bool) {
		if !bx.AddrPort.IsValid() {
			xlog.Errata.Println("no DAP")
		} else if n, err := bx.Tx(conn); err != nil {
			xlog.Errata.Printf("tx %v <- %v: %v",
				bx.AddrPort, la, err)
		} else {
			xlog.Trace.Printf("tx[%d] %v <- %v",
				n, bx.AddrPort, la)
			ok = true
		}
		bx.Return()
		return
	})
}
