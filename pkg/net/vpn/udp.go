// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
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

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box"
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
	verbose.Printf("start %v rx routine", udp.LocalAddr())
	defer verbose.Printf("stopped %v rx routine", udp.LocalAddr())
	udp.SetReadDeadline(time.Time{})
	for dur := minPktRxDeadline; ctx.Err() == nil; {
		err := udp.SetReadDeadline(time.Now().Add(dur))
		if err != nil {
			verbose.Print(err)
			break
		} else if bx, err := box.NewRx(udp); err == nil {
			verbose.Printf("rx %d bytes from %v",
				bx.Len(), bx.AddrPort)
			ch <- bx
		} else if operr, ok := err.(*net.OpError); ok {
			if !operr.Timeout() {
				verbose.Print(err)
				break
			}
		} else if !errors.Is(err, os.ErrDeadlineExceeded) {
			verbose.Print(err)
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
	verbose.Printf("start %v tx routine", la)
	defer verbose.Printf("stopped %v tx routine", la)
	for {
		select {
		case <-ctx.Done():
			return
		case bx, ok := <-ch:
			if !ok {
				verbose.Println("pkt tx ch closed")
				return
			}
			ta := bx.AddrPort
			n, err := bx.Tx(udp)
			if err != nil {
				verbose.Printf("tx from %v to %v: %v",
					la, ta, err)
			} else {
				verbose.Printf("tx %d bytes from %v to %v",
					n, la, ta)
			}
			bx.Return()
		}
	}
}
