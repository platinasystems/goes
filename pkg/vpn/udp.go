// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"errors"
	"net"
	"os"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/xlog"
)

const (
	minPktRxDeadline = 10 * time.Millisecond
	maxPktRxDeadline = 250 * time.Millisecond
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
	udp.SetReadDeadline(time.Time{})
	for dur := minPktRxDeadline; ctx.Err() == nil; {
		err := udp.SetReadDeadline(time.Now().Add(dur))
		if err != nil {
			xlog.Errata.Print(err)
			break
		} else if bx, err := box.NewRx(udp); err == nil {
			xlog.Trace.Printf("rx %d bytes from %v",
				bx.Len(), bx.AddrPort)
			ch <- bx
		} else if operr, ok := err.(*net.OpError); ok {
			if !operr.Timeout() {
				xlog.Errata.Print(err)
				break
			}
		} else if !errors.Is(err, os.ErrDeadlineExceeded) {
			xlog.Errata.Print(err)
			break
		} else if dur < maxPktRxDeadline {
			if dur *= 2; dur > maxPktRxDeadline {
				dur = maxPktRxDeadline
			}
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
			} else if n, err := bx.Tx(udp); err != nil {
				xlog.Info.Printf("tx from %v to %v: %v",
					la, tap, err)
			} else {
				xlog.Trace.Printf("tx %d bytes from %v to %v",
					n, la, tap)
			}
			bx.Return()
		}
	}
}
