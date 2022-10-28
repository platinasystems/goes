// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bridge

import (
	"context"
	"crypto/tls"
	"os"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
)

type Bridge struct {
	input chan *input
	join  chan *tls.Conn
	leave chan *tls.Conn
}

func New(ctx context.Context, wg *sync.WaitGroup) *Bridge {
	br := &Bridge{
		input: make(chan *input, 16),
		join:  make(chan *tls.Conn, 4),
		leave: make(chan *tls.Conn, 4),
	}
	wg.Add(1)
	go br.forward(ctx, wg)
	return br
}

func (br *Bridge) Join(ctx context.Context, conn *tls.Conn) error {
	br.join <- conn
	defer func() { br.leave <- conn }()

	dec := lv.NewDecoder(poll.With(ctx, conn))
	for {
		in := pool.Get().(*input)
		in.c = conn
		n, err := dec.Read(in.b)
		if err != nil {
			in.free()
			return err
		}
		in.b = in.b[:n]
		br.input <- in
	}
	return nil
}

func (br *Bridge) forward(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	var flood []*tls.Conn
	learned := make(map[uint64]*tls.Conn)

	learn := func(b []byte, c *tls.Conn) {
		learned[ea64(b[6:])] = c
	}
	forward := func(ctx context.Context, c *tls.Conn, b []byte) error {
		_, err := lv.NewEncoder(write.With(ctx, c)).Write(b)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			flood = flood[:0]
			for k := range learned {
				delete(learned, k)
			}
			return
		case in := <-br.input:
			if len(in.b) < 14 {
				in.free()
				continue
			}
			if (in.b[6] & 1) == 0 {
				learn(in.b, in.c)
			}
			if (in.b[0] & 1) == 0 {
				da := ea64(in.b)
				if c, ok := learned[da]; ok {
					err := forward(ctx, c, in.b)
					if err == nil {
						in.free()
						continue
					}
					delete(learned, da)
				}
			}
			for _, c := range flood {
				forward(ctx, c, in.b)
			}
			in.free()
		case c := <-br.join:
			flood = append(flood, c)
		case c := <-br.leave:
			for i, v := range flood {
				if v == c {
					copy(flood[i:], flood[i+1:])
					flood = flood[:len(flood)-1]
				}
			}
		}
	}
}

type input struct {
	c *tls.Conn
	b []byte
}

var (
	size = os.Getpagesize()
	pool = sync.Pool{
		New: func() any {
			return &input{
				b: make([]byte, size, size),
			}
		},
	}
)

func (in *input) free() {
	if cap(in.b) == size {
		in.c = nil
		in.b = in.b[:size]
		pool.Put(in)
	}
}

func ea64(b []byte) uint64 {
	ea := uint64(b[0]) << 40
	ea |= uint64(b[1]) << 32
	ea |= uint64(b[2]) << 24
	ea |= uint64(b[3]) << 16
	ea |= uint64(b[4]) << 8
	ea |= uint64(b[5])
	return ea
}
