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
)

const (
	minPktRxDeadline = 10 * time.Millisecond
	maxPktRxDeadline = 250 * time.Millisecond
)

func pktRxRoutine(
	ctx context.Context,
	wg *sync.WaitGroup,
	udp *net.UDPConn,
	ch chan<- *crate,
) {
	var n int
	defer wg.Done()
	defer close(ch)
	verbose.Print("start ", udp.LocalAddr(), " rx routine")
	defer verbose.Print("stopped ", udp.LocalAddr(), " rx routine")
	udp.SetReadDeadline(time.Time{})
	for dur := minPktRxDeadline; ctx.Err() == nil; {
		err := udp.SetReadDeadline(time.Now().Add(dur))
		if err != nil {
			verbose.Print(err)
			break
		}
		crate := newCrate()
		n, crate.ap, err = udp.ReadFromUDPAddrPort(crate.box)
		if err == nil {
			crate.box = crate.box[:n]
			verbose.Println("rx", n, "bytes from", crate.ap)
			ch <- crate
		} else if operr, ok := err.(*net.OpError); ok {
			if !operr.Timeout() {
				verbose.Print(err)
				break
			}
		} else if !errors.Is(err, os.ErrDeadlineExceeded) {
			verbose.Print(err)
			break
		}
		if dur < maxPktRxDeadline {
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
	ch <-chan *crate,
) {
	defer wg.Done()
	la := udp.LocalAddr()
	verbose.Print("start ", la, " tx routine")
	defer verbose.Print("stopped ", la, " tx routine")
	for {
		select {
		case <-ctx.Done():
			return
		case crate := <-ch:
			udp.WriteToUDPAddrPort(crate.box, crate.ap)
			verbose.Printf("tx %d bytes from %v to %v",
				len(crate.box), la, crate.ap)
			inventory.Put(crate)
		}
	}
}
