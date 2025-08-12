// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/platinasystems/goes/v2/pkg/xsync"
)

const MsgTestMTU = 1500

func BenchmarkMsgPool(b *testing.B) {
	mp := NewMsgPool(MsgTestMTU)
	b.SetBytes(MsgTestMTU)
	for b.Loop() {
		mp.Put(mp.Get())
	}
}

func BenchmarkMsgChan(b *testing.B) {
	mp := NewMsgPool(MsgTestMTU)
	b.SetBytes(MsgTestMTU)
	for _, depth := range []int{
		0, 1, 2, 4, 8, 16, 32, 64,
	} {
		b.Run(fmt.Sprint("depth-", depth), func(b *testing.B) {
			var wg xsync.WaitGroup
			var dropped uint32

			b.SetBytes(MsgTestMTU)

			ch := make(chan *Msg, depth)

			wg.Go(func() {
				var got, want uint32
				for m := range ch {
					got = Uint32(m.Data)
					mp.Put(m)
					if got > want {
						dropped = got - want
					}
					want = got + 1
				}
			})
			for i := uint32(0); b.Loop(); i++ {
				m := mp.Get()
				Encode32(m.Data, i)
				ch <- m
			}
			close(ch)
			wg.Wait()
			if dropped != 0 {
				b.Errorf("dropped %d/%d (%%%g)", dropped, b.N,
					100*float32(dropped)/float32(b.N))
			}
		})
	}
}

func BenchmarkMsgService(b *testing.B) {
	mp := NewMsgPool(MsgTestMTU)
	benchmarkMsgService(b, mp.RecvService, mp.SendService)
}

func BenchmarkMsgBatchService(b *testing.B) {
	if !CanBatch {
		b.Skip()
	}
	mp := NewMsgPool(MsgTestMTU)
	benchmarkMsgService(b, mp.RecvBatchService, mp.SendBatchService)
}

func benchmarkMsgService(
	b *testing.B,
	recvsvc func(context.Context, chan<- *Msg, net.PacketConn) error,
	sendsvc func(net.PacketConn, <-chan *Msg) error,
) {
	const address = "127.0.0.1:9000"
	mp := NewMsgPool(MsgTestMTU)
	for _, depth := range []int{
		1, 4, 8, 16, 64,
	} {
		b.Run(fmt.Sprint("depth-", depth), func(b *testing.B) {
			var wg xsync.WaitGroup
			var dropped uint32

			ctx, cancel := context.WithCancel(b.Context())

			b.SetBytes(MsgTestMTU)

			rconn, err := net.ListenPacket("udp", address)
			if err != nil {
				b.Error(err)
				return
			}
			defer rconn.Close()

			sconn, err := DialPacket("udp", address)
			if err != nil {
				b.Error(err)
				return
			}
			defer sconn.Close()

			rch := make(chan *Msg, depth)
			sch := make(chan *Msg, depth)

			wg.Go(func() {
				if err := recvsvc(ctx, rch, rconn); err != nil {
					b.Fatal("rcvsvc:", err)
				}
			})
			wg.Go(func() {
				if err := sendsvc(sconn, sch); err != nil {
					b.Fatal("sndsvc:", err)
				}
			})
			wg.Go(func() {
				var got, want uint32
				for m := range rch {
					got = Uint32(m.Data)
					mp.Put(m)
					if got > want {
						dropped = got - want
					}
					want = got + 1
				}
			})

			for i := uint32(0); b.Loop(); i++ {
				m := mp.Get()
				Encode32(m.Data, i)
				sch <- m
			}

			close(sch)
			cancel()
			wg.Wait()

			if dropped != 0 {
				b.Logf("dropped %d/%d (%%%g)", dropped, b.N,
					100*float32(dropped)/float32(b.N))
			}
		})
	}
}
