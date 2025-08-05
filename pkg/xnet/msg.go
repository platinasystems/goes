// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import (
	"context"
	"errors"
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

type Msg struct {
	netip.AddrPort
	Data []byte
}

type MsgPool struct {
	mtu int
	p   *sync.Pool
}

const BatchCap = 8

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

// Copy pooled messages from socket to channel until socket is closed.
func (mp *MsgPool) RecvService(ch chan<- *Msg, conn net.PacketConn) error {
	return mp.rcvsvc(ch, conn)
}

// Copy messages from channel to socket and return to pool until channel is
// closed or write error.
func (mp *MsgPool) SendService(conn net.PacketConn, ch <-chan *Msg) error {
	return mp.sndsvc(conn, ch)
}

// Copy messages from channel and return to pool until channel.
func (mp *MsgPool) WriteService(w io.WriteCloser, ch <-chan *Msg) error {
	defer w.Close()
	for m := range ch {
		_, err := w.Write(m.Data)
		mp.Put(m)
		if err != nil {
			return xerrors.Suppress(err, net.ErrClosed)
		}
	}
	return nil
}
