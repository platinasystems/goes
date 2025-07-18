// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"net"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

var (
	udp  *net.UDPConn
	udpC chan *xnet.Msg
)

func udpInit(port uint16) (err error) {
	udpC = make(chan *xnet.Msg, 1)
	udp, err = net.ListenUDP("udp", &net.UDPAddr{
		Port: int(port),
	})
	return
}

func udpStream() {
	la := udp.LocalAddr()
	xlog.Trace.Print("start ", la.Network(), ":", la)
	err := xerrors.Suppress(mp.StreamUDP(udpC, udp), net.ErrClosed)
	if err != nil {
		xlog.Errata.Print("quit ", la.Network(), ":", la)
	} else {
		xlog.Trace.Print("stopped ", la.Network(), ":", la)
	}
}
