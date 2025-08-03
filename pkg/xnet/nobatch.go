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

func (mp *MsgPool) rcvsvc(ch chan<- *Msg, sock *net.UDPConn) error {
	var (
		err error
		n   int
	)

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
		} else if errors.Is(err, net.ErrClosed) {
			return nil
		} else {
			return err
		}
	}
}

func (mp *MsgPool) sndsvc(sock *net.UDPConn, ch <-chan *Msg) error {
	var err error

	defer sock.Close()

	for m := range ch {
		if !m.AddrPort.IsValid() {
			return xerrors.Invalid("address")
		} else if m.AddrPort.Port() == 0 {
			return xerrors.Invalid("port")
		} else {
			_, err = sock.WriteToUDPAddrPort(m.Data, m.AddrPort)
		}
		mp.Put(m)
		if err != nil {
			return xerrors.Suppress(err, net.ErrClosed)
		}
	}
	return nil
}
