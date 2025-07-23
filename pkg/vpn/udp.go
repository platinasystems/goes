// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"errors"
	"net"
	"net/netip"
	"runtime"

	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

// Open socket, make channels and start routines.
// If not error, close the returned write channel when done.
func startUDP(port uint16) (<-chan *xnet.Msg, chan<- *xnet.Msg, error) {
	fromC := make(chan *xnet.Msg, SizeofFromVpnC)
	toC := make(chan *xnet.Msg, SizeofToVpnC)
	sock, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: int(port),
	})
	if err != nil {
		return nil, nil, err
	}
	wg.Go(func() { streamFromUDP(fromC, sock) })
	wg.Go(func() { streamToUDP(sock, toC) })
	return fromC, toC, err
}

// Forward messages from socket to channel until not [xos.IsBlocked] error;
// then close channel.
func streamFromUDP(ch chan<- *xnet.Msg, sock *net.UDPConn) {
	var (
		err error
		n   int
	)

	la := sock.LocalAddr()
	xlog.Trace.Print("start udp:", la, " read")
	defer close(ch)

	m := mp.Get()
	defer mp.Put(m)

	for {
		n, m.AddrPort, err = sock.ReadFromUDPAddrPort(m.Data)
		if err == nil {
			if addr := m.AddrPort.Addr(); addr.Is4In6() {
				addr = addr.Unmap()
				port := m.AddrPort.Port()
				m.AddrPort = netip.AddrPortFrom(addr, port)
			}
			m.Data = m.Data[:n]
			ch <- m
			m = mp.Get()
		} else if xos.IsBlocked(err) {
			runtime.Gosched()
		} else {
			break
		}
	}
	if errors.Is(err, net.ErrClosed) {
		xlog.Trace.Print("stopped udp:", la, " read")
	} else {
		xlog.Errata.Print("quit udp:", la, " read: ", err)
	}
}

func streamToUDP(sock *net.UDPConn, ch <-chan *xnet.Msg) {
	var err error

	la := sock.LocalAddr()
	xlog.Trace.Print("start udp:", la, " write")
	defer sock.Close()

	for m := range ch {
		if !m.AddrPort.IsValid() {
			xlog.Errata.Print("drop udp:", la, " missing Addr")
		} else if m.AddrPort.Port() == 0 {
			xlog.Errata.Print("drop udp:", la, " missing Port")
		} else {
			_, err = sock.WriteToUDPAddrPort(m.Data, m.AddrPort)
		}
		mp.Put(m)
		if err != nil {
			break
		}
	}
	if err != nil {
		xlog.Errata.Print("quit udp:", la, " write: ", err)
	} else {
		xlog.Trace.Print("stopped udp:", la, " write")
	}
}
