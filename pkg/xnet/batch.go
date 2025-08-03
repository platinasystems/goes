// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build linux

package xnet

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"runtime"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xos"
	"golang.org/x/net/ipv4"
	"golang.org/x/sys/unix"
)

func (mp *MsgPool) rcvsvc(ch chan<- *Msg, sock *net.UDPConn) error {
	defer close(ch)
	xsock := ipv4.NewPacketConn(sock)
	msgs := make([]*Msg, BatchCap, BatchCap)
	batch := make([]ipv4.Message, BatchCap, BatchCap)
	for i := range msgs {
		msgs[i] = mp.Get()
		batch[i].Buffers = make([][]byte, 1)
		batch[i].Buffers[0] = msgs[i].Data
	}
	var flags int
	if runtime.NumCPU() > 1 {
		flags = unix.MSG_WAITFORONE
	} else {
		flags = unix.MSG_DONTWAIT
	}
	for {
		n, err := xsock.ReadBatch(batch, flags)
		if err != nil {
			if xos.IsBlocked(err) {
				runtime.Gosched()
				continue
			} else if errors.Is(err, net.ErrClosed) {
				return nil
			} else {
				return err
			}
		}
		if n == 0 {
			runtime.Gosched()
			continue
		}
		for i := 0; i < n; i++ {
			ia := batch[i].Addr
			if a, ok := ia.(*net.UDPAddr); !ok {
				return fmt.Errorf("!udp: %T{%v}", ia, ia)
			} else if aa, ok := netip.AddrFromSlice(a.IP); !ok {
				return fmt.Errorf("can't convert: %v", a.IP)
			} else {
				msgs[i].AddrPort = netip.
					AddrPortFrom(aa, uint16(a.Port))
				msgs[i].Data = msgs[i].Data[:batch[i].N]
				ch <- msgs[i]
				msgs[i] = mp.Get()
				batch[i].Buffers[0] = msgs[i].Data
			}
		}
	}
}

func (mp *MsgPool) sndsvc(sock *net.UDPConn, ch <-chan *Msg) error {
	defer sock.Close()

	xsock := ipv4.NewPacketConn(sock)
	msgs := make([]*Msg, BatchCap, BatchCap)
	batch := make([]ipv4.Message, BatchCap, BatchCap)
	for i := range batch {
		batch[i].Buffers = make([][]byte, 1)
	}
	defer func() {
		for i, m := range msgs {
			if m != nil {
				mp.Put(m)
				msgs[i] = nil
			}
			batch[i].Buffers[0] = nil
		}
	}()
	for {
		var i int
	selection:
		for i = 0; i < BatchCap; i++ {
			select {
			case m, ok := <-ch:
				if !ok {
					return nil
				}
				msgs[i] = m
				batch[i].Addr = netip2UDPAddr(m.AddrPort)
				batch[i].Buffers[0] = m.Data
			default:
				break selection
			}
		}
		for tn := 0; tn < i; {
			n, err := xsock.WriteBatch(batch[tn:i], 0)
			if err != nil {
				return xerrors.Suppress(err, net.ErrClosed)
			}
			tn += n
		}
		for i -= 1; i >= 0; i-- {
			mp.Put(msgs[i])
			msgs[i] = nil
		}
	}
}

func netip2UDPAddr(ap netip.AddrPort) *net.UDPAddr {
	var ip net.IP
	var zone string
	if a := ap.Addr(); a.Is4() {
		a4 := a.As4()
		ip = net.IP(a4[:])
	} else {
		a16 := a.As16()
		ip = net.IP(a16[:])
		// FIXME on guests, set zone to interface name (e.g. "tun0")
	}
	return &net.UDPAddr{IP: ip, Port: int(ap.Port()), Zone: zone}
}
