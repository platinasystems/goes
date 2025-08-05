// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !linux

package xnet

import (
	"errors"
	"net"
	"net/netip"
	"runtime"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

type ReadFromUDPAddrPorter interface {
	ReadFromUDPAddrPort([]byte) (int, netip.AddrPort, error)
}

type WriteToUDPAddrPorter interface {
	WriteToUDPAddrPort([]byte, netip.AddrPort) (int, error)
}

func (mp *MsgPool) rcvsvc(ch chan<- *Msg, conn net.PacketConn) error {
	var (
		err error
		n   int
	)
	rcv := func(b []byte) (int, netip.AddrPort, error) {
		var ap netip.AddrPort
		n, addr, err := conn.ReadFrom(b)
		if err == nil {
			ap, err = netip.ParseAddrPort(addr.String())
		}
		return n, ap, err
	}
	if udp, ok := conn.(ReadFromUDPAddrPorter); ok {
		rcv = udp.ReadFromUDPAddrPort
	}

	defer close(ch)

	m := mp.Get()
	defer mp.Put(m)

	for {
		n, m.AddrPort, err = rcv(m.Data)
		if err == nil {
			if m.AddrPort.IsValid() {
				if addr := m.AddrPort.Addr(); addr.Is4In6() {
					addr = addr.Unmap()
					port := m.AddrPort.Port()
					m.AddrPort = netip.
						AddrPortFrom(addr, port)
				}
			}
			m.Data = m.Data[:n]
			ch <- m
			m = mp.Get()
		} else if xos.IsBlocked(err) {
			runtime.Gosched()
		} else if errors.Is(err, net.ErrClosed) {
			return nil
		} else {
			return err
		}
	}
}

func (mp *MsgPool) sndsvc(conn net.PacketConn, ch <-chan *Msg) error {
	snd := func(b []byte, ap netip.AddrPort) (int, error) {
		addr := net.UDPAddrFromAddrPort(ap)
		return conn.WriteTo(b, addr)
	}
	if udp, ok := conn.(WriteToUDPAddrPorter); ok {
		snd = udp.WriteToUDPAddrPort
	}

	defer conn.Close()

	for m := range ch {
		_, err := snd(m.Data, m.AddrPort)
		mp.Put(m)
		if err != nil {
			return xerrors.Suppress(err, net.ErrClosed)
		}
	}
	return nil
}
