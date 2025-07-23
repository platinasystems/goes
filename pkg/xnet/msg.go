// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import (
	"context"
	"net"
	"net/netip"
	"sync"
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
