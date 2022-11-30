// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bridge

import (
	"context"
	"crypto/tls"
	"errors"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/frame"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

var ErrTooShort = errors.New("len(frame) < 14")

type Bridge struct {
	name  string
	input chan *input
	join  chan *tls.Conn
	leave chan *tls.Conn
}

func New(ctx context.Context, wg *sync.WaitGroup, name string) *Bridge {
	br := &Bridge{
		name:  name,
		input: make(chan *input, 16),
		join:  make(chan *tls.Conn, 4),
		leave: make(chan *tls.Conn, 4),
	}
	wg.Add(1)
	go br.forward(ctx, wg)
	return br
}

func (br *Bridge) Join(ctx context.Context, c *tls.Conn) {
	br.join <- c
	defer func() { br.leave <- c }()

	sub := dns0(c)
	dec := lv.NewDecoder(poll.With(ctx, c))
	for {
		pg := page.New()
		n, err := dec.Read(pg)
		if err != nil {
			page.Free(pg)
			style.Error(err)
			break
		}
		if n < 14 {
			page.Free(pg)
			style.Error(ErrTooShort)
			break
		}
		in := newinput(c, pg[:n])
		style.Println(br.name, "<-", sub, frame.Eth(in.pg))
		br.input <- in
	}
}

func (br *Bridge) forward(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	var flood []*tls.Conn
	lookup := make(map[uint64]*tls.Conn)

	for {
		select {
		case <-ctx.Done():
			flood = flood[:0]
			for k := range lookup {
				delete(lookup, k)
			}
			return
		case c := <-br.join:
			flood = append(flood, c)
		case c := <-br.leave:
			for i, v := range flood {
				if v == c {
					copy(flood[i:], flood[i+1:])
					flood = flood[:len(flood)-1]
				}
			}
			for k, v := range lookup {
				if v == c {
					delete(lookup, k)
					break
				}
			}
		case in := <-br.input:
			eth := frame.Eth(in.pg)
			if eth.ShouldLearn() {
				lookup[eth.SA()] = in.c
			}
			da := eth.DA()
			if c, ok := lookup[da]; ok {
				br.send(ctx, c, in)
			} else {
				for _, c := range flood {
					if c != in.c {
						br.send(ctx, c, in)
					}
				}
			}
			in.free()
		}
	}
}

func (br *Bridge) send(ctx context.Context, c *tls.Conn, in *input) {
	_, err := lv.NewEncoder(write.With(ctx, c)).Write(in.pg)
	if err != nil {
		style.Errorln(br.name, "->", dns0(c), err)
	} else {
		style.Println(br.name, "->", dns0(c), frame.Eth(in.pg))
	}
}

type input struct {
	c  *tls.Conn
	pg []byte
}

var pool = sync.Pool{New: func() any { return new(input) }}

func newinput(c *tls.Conn, pg []byte) *input {
	in := pool.Get().(*input)
	in.c = c
	in.pg = pg
	return in
}

func (in *input) free() {
	page.Free(in.pg)
	in.c = nil
	in.pg = in.pg[:0]
	pool.Put(in)
}

func dns0(c *tls.Conn) string {
	if cs := c.ConnectionState(); len(cs.PeerCertificates) > 0 &&
		len(cs.PeerCertificates[0].DNSNames) > 0 {
		return cs.PeerCertificates[0].DNSNames[0]
	}
	return "anonymous"
}
