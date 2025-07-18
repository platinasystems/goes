// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import (
	"context"
	"io"
	"net"
	"net/netip"
	"runtime"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xos"
)

var zap netip.AddrPort

type AddrPorter interface {
	AddrPort() netip.AddrPort
}

type RemoteAddrer interface {
	RemoteAddr() net.Addr
}

type Msg struct {
	netip.AddrPort
	Data []byte
}

type MsgPool struct {
	mtu int
	p   *sync.Pool
}

func NewMsgPool(mtu int) *MsgPool {
	return &MsgPool{
		mtu: mtu,
		p: &sync.Pool{
			New: func() any {
				return &Msg{
					AddrPort: zap,
					Data:     make([]byte, mtu, mtu),
				}
			},
		},
	}
}

func (mp *MsgPool) Get() *Msg {
	return mp.p.Get().(*Msg)
}

func (mp *MsgPool) Put(m *Msg) {
	if cap(m.Data) == mp.mtu {
		m.AddrPort = zap
		m.Data = m.Data[:mp.mtu]
		mp.p.Put(m)
	} else {
		// Let GC deal with this.
		m.Data = m.Data[:0]
		m = nil
	}
}

// Discard message and return false if context is done before message is
// channeled; otherwise return true.
func (mp *MsgPool) Queue(ctx context.Context, ch chan<- *Msg, m *Msg) bool {
	select {
	case <-ctx.Done():
		mp.Put(m)
	case ch <- m:
		return true
	}
	return false
}

// Forward messages from reader to channel until not [xos.IsBlocked] error;
// then close channel.
func (mp *MsgPool) StreamReader(ch chan<- *Msg, r io.Reader) error {
	var (
		err error
		n   int
		ra  netip.AddrPort
	)
	defer close(ch)
	m := mp.Get()
	defer mp.Put(m)
	if raer, ok := r.(RemoteAddrer); ok {
		if aper, ok := raer.RemoteAddr().(AddrPorter); ok {
			ra = aper.AddrPort()
		}
	}
	for {
		n, err = r.Read(m.Data)
		if err == nil {
			m.AddrPort = ra
			m.Data = m.Data[:n]
			ch <- m
			m = mp.Get()
		} else if xos.IsBlocked(err) {
			runtime.Gosched()
		} else {
			return err
		}
	}
}

// Forward messages from socket to channel until not [xos.IsBlocked] error;
// then close channel.
func (mp *MsgPool) StreamUDP(ch chan<- *Msg, udp *net.UDPConn) error {
	var (
		err error
		n   int
	)
	defer close(ch)
	m := mp.Get()
	defer mp.Put(m)
	for {
		n, m.AddrPort, err = udp.ReadFromUDPAddrPort(m.Data)
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
			return err
		}
	}
}
