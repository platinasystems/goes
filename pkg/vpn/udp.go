// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"net"

	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

// Open socket, make channels and start packet read/write routines.
// If not error, close the returned write channel when done.
func startUDP(port uint16) (<-chan *xnet.Msg, chan<- *xnet.Msg, error) {
	fromC := make(chan *xnet.Msg, FromVpnCap)
	toC := make(chan *xnet.Msg, ToVpnCap)
	sock, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: int(port),
	})
	if err != nil {
		return nil, nil, err
	}
	wg.Go(func() {
		xlog.Trace.Println("start receive service")
		err := mp.RecvService(fromC, sock)
		if err != nil {
			xlog.Errata.Println("quit receive service:", err)
		} else {
			xlog.Trace.Println("stopped receive service")
		}
	})
	wg.Go(func() {
		xlog.Trace.Println("start send service")
		err := mp.SendService(sock, toC)
		if err != nil {
			xlog.Errata.Println("quit send service:", err)
		} else {
			xlog.Trace.Println("stopped send service")
		}
	})
	return fromC, toC, err
}
