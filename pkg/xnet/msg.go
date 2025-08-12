// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"runtime"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

var zap netip.AddrPort

type AddrPorter interface {
	AddrPort() netip.AddrPort
}

type RemoteAddrer interface {
	RemoteAddr() net.Addr
}

func DialPacket(network, address string) (net.PacketConn, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	pktconn, ok := conn.(net.PacketConn)
	if !ok {
		conn.Close()
		return nil, fmt.Errorf("%T isn't PacketConn", conn)
	}
	return pktconn, nil
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

func (mp *MsgPool) ReadService(ch chan<- *Msg, r io.Reader) error {
	m := mp.Get()
	defer mp.Put(m)
	defer close(ch)
	for {
		n, err := r.Read(m.Data)
		if err == nil {
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

// Copy to pooled messages from socket then send to channel
// until context is done; then close channel before returning.
func (mp *MsgPool) RecvService(
	ctx context.Context, ch chan<- *Msg, conn net.PacketConn,
) error {
	var err error

	defer close(ch)

	var rcv func([]byte) (int, netip.AddrPort, error)
	if udp, ok := conn.(*net.UDPConn); ok {
		if ra := udp.RemoteAddr(); ra == nil {
			rcv = udp.ReadFromUDPAddrPort
		} else {
			var rap netip.AddrPort
			if udpa, ok := ra.(*net.UDPAddr); ok {
				a, _ := netip.AddrFromSlice(udpa.IP)
				rap = netip.AddrPortFrom(a, uint16(udpa.Port))
			} else {
				rap, err = netip.ParseAddrPort(udpa.String())
				if err != nil {
					return err
				}
			}
			rcv = func(b []byte) (int, netip.AddrPort, error) {
				n, err := udp.Read(b)
				return n, rap, err
			}
		}
	} else {
		rcv = func(b []byte) (int, netip.AddrPort, error) {
			var ap netip.AddrPort
			n, addr, err := conn.ReadFrom(b)
			if err == nil {
				ap, err = netip.ParseAddrPort(addr.String())
			}
			return n, ap, err
		}
	}

	m := mp.Get()
	defer mp.Put(m)

	for ctx.Err() == nil {
		n, ap, err := rcv(m.Data)
		if err != nil {
			if xos.IsBlocked(err) {
				runtime.Gosched()
				continue
			}
			return xerrors.Suppress(err, net.ErrClosed)
		}
		if !ap.IsValid() {
			return xerrors.Invalid("address")
		}
		if addr := ap.Addr(); addr.Is4In6() {
			addr = addr.Unmap()
			port := m.AddrPort.Port()
			m.AddrPort = netip.AddrPortFrom(addr, port)
		} else {
			m.AddrPort = ap
		}
		m.Data = m.Data[:n]
		ch <- m
		m = mp.Get()
	}
	return nil
}

// Copy messages from channel to socket and return to pool until the channel is
// closed.
func (mp *MsgPool) SendService(conn net.PacketConn, ch <-chan *Msg) error {
	var err error

	var snd func([]byte, netip.AddrPort) (int, error)
	if udp, ok := conn.(*net.UDPConn); ok {
		if udp.RemoteAddr() != nil {
			snd = func(b []byte, _ netip.AddrPort) (int, error) {
				return udp.Write(b)
			}
		} else {
			snd = udp.WriteToUDPAddrPort
		}
	} else {
		snd = func(b []byte, ap netip.AddrPort) (int, error) {
			addr := net.UDPAddrFromAddrPort(ap)
			return conn.WriteTo(b, addr)
		}
	}

	for m := range ch {
		_, err = snd(m.Data, m.AddrPort)
		mp.Put(m)
		if err != nil {
			break
		}
	}
	return xerrors.Suppress(err, net.ErrClosed)
}

// Copy messages from channel and return to pool until channel.
func (mp *MsgPool) WriteService(w io.Writer, ch <-chan *Msg) error {
	for m := range ch {
		_, err := w.Write(m.Data)
		mp.Put(m)
		if err != nil {
			return xerrors.Suppress(err, net.ErrClosed)
		}
	}
	return nil
}
