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

// Forward messages from reader to channel until EOF or context is done;
// then close channel.
func (mp *MsgPool) ReadMsgs(
	ctx context.Context, r io.Reader, ch chan<- *Msg,
) error {
	defer close(ch)
	m := mp.Get()
	defer mp.Put(m)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		n, err := r.Read(m.Data)
		if err != nil {
			if !xos.IsBlocked(err) {
				runtime.Gosched()
				continue
			}
			return err
		}
		m.Data = m.Data[:n]
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- m:
			m = mp.Get()
		}
	}
}

// Forward messages from connection to channel until the connection is closed
// or context is done; then close channel.
func (mp *MsgPool) RecvMsgs(
	ctx context.Context, conn *net.UDPConn, ch chan<- *Msg,
) error {
	var n int
	var err error
	defer close(ch)
	m := mp.Get()
	defer mp.Put(m)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		n, m.AddrPort, err = conn.ReadFromUDPAddrPort(m.Data)
		if err != nil {
			return err
		}
		m.Data = m.Data[:n]
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- m:
			m = mp.Get()
		}
	}
}
