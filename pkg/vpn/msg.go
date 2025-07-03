// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

var (
	mp  *xnet.MsgPool
	udp *net.UDPConn
)

func udpLocalAddrPort(udp *net.UDPConn) (netip.AddrPort, error) {
	s := udp.LocalAddr().String()
	return xerrors.MarkResult(netip.ParseAddrPort(s))
}

func udpListen() (err error) {
	mp = xnet.NewMsgPool(netph.ETHMTU)

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
