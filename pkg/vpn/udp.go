// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"net"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/xlog"
)

func pktRxRoutine(
	ctx context.Context,
	wg *sync.WaitGroup,
	udp *net.UDPConn,
	ch chan<- *box.Box,
) {
	defer wg.Done()
	defer close(ch)
	la := udp.LocalAddr()
	xlog.Info.Printf("start %v rx routine", la)
	defer xlog.Info.Printf("stopped %v rx routine", la)
	for ctx.Err() == nil {
		if bx, err := box.NewRx(ctx, udp); err != nil {
			xlog.Errata.Print(err)
			break
		} else if bx != nil {
			xlog.Trace.Printf("rx %d bytes from %v",
				bx.Len(), bx.AddrPort)
			bx.Queue(ctx, ch)
		}
	}
}

func pktTxRoutine(
	ctx context.Context,
	wg *sync.WaitGroup,
	udp *net.UDPConn,
	ch <-chan *box.Box,
) {
	defer wg.Done()
	la := udp.LocalAddr()
	xlog.Info.Printf("start %v tx routine", la)
	defer xlog.Info.Printf("stopped %v tx routine", la)
	for {
		select {
		case <-ctx.Done():
			return
		case bx, ok := <-ch:
			if !ok {
				xlog.Info.Println("pkt tx ch closed")
				return
			}
			if tap := bx.AddrPort; !tap.IsValid() {
				xlog.Errata.Println("no DAP")
			} else if n, err := bx.Tx(ctx, udp); err != nil {
				xlog.Errata.Printf("tx from %v to %v: %v",
					la, tap, err)
			} else {
				xlog.Trace.Printf("tx %d bytes from %v to %v",
					n, la, tap)
			}
			bx.Return()
		}
	}
}
