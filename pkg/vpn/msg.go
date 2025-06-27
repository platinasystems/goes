// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"io"
	"net"
	"net/netip"
	"runtime"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

var (
	mp  *xnet.MsgPool
	udp *net.UDPConn
	rmc chan *xnet.Msg
)

func udpLocalAddrPort(udp *net.UDPConn) (netip.AddrPort, error) {
	s := udp.LocalAddr().String()
	return xerrors.MarkResult(netip.ParseAddrPort(s))
}

func udpListen() (err error) {
	mp = xnet.NewMsgPool(netph.ETHMTU)

	if runtime.NumCPU() > 1 {
		rmc = make(chan *xnet.Msg, 4)
	} else {
		rmc = make(chan *xnet.Msg)
	}

	udpnet := "udp"
	if a := vpnListen.Addr(); a.Is4() {
		udpnet = "udp4"
	} else if a.Is6() {
		udpnet = "udp6"
	}
	udp, err = xerrors.MarkResult(net.ListenUDP(udpnet, &net.UDPAddr{
		IP:   vpnListen.Addr().AsSlice(),
		Port: int(vpnListen.Port()),
	}))
	return
}

func udpSend(ctx context.Context, m *xnet.Msg) error {
	_, err := udp.WriteToUDPAddrPort(m.Data, m.AddrPort)
	return err
}

func udpStream(ctx context.Context) error {
	la := udp.LocalAddr()
	xlog.Info.Println("start stream from", la)
	defer xlog.Info.Println("stopped stream from", la)
	err := mp.RecvMsgs(ctx, udp, rmc)
	return xerrors.Suppress(err, net.ErrClosed, context.Canceled, io.EOF)
}
