// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build ignore

package tlsx

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/frame"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

type bridgeInput struct {
	c  *tls.Conn
	pg []byte
}

var bridge = struct {
	name    string
	inputch chan *bridgeInput
	joinch  chan *tls.Conn
	leavech chan *tls.Conn
	pool    sync.Pool
}{
	inputch: make(chan *bridgeInput, 16),
	joinch:  make(chan *tls.Conn, 4),
	leavech: make(chan *tls.Conn, 4),
	pool:    sync.Pool{New: func() any { return new(bridgeInput) }},
}

func JoinBridge(ctx context.Context, c *tls.Conn, args []string) {
	var tenant string

	self := Self().SKI()
	ra := remoteAddr(c)
	dec := lv.NewDecoder(poll.WithReader(ctx, c))
	enc := lv.NewEncoder(write.With(ctx, c))

	if len(bridge.name) == 0 {
		enc.Encode(ErrUnavailable)
		return
	}

	if cs := c.ConnectionState(); len(cs.PeerCertificates) == 0 {
		enc.Encode(fmt.Errorf("%v: no certficate", ra))
		return
	} else if len(cs.PeerCertificates[0].DNSNames) == 0 {
		enc.Encode(fmt.Errorf("%v: unnamed", ra))
		return
	} else {
		tenant = cs.PeerCertificates[0].DNSNames[0]
		ski := hex.EncodeToString(cs.PeerCertificates[0].SubjectKeyId)
		if ski == self {
			tenant += fmt.Sprint("@", ra)
		}
	}
	if p, err := leaseContract(tenant, args...); err != nil {
		enc.Encode(err)
		return
	} else {
		enc.Encode(p.String(), nil)
	}

	bridge.joinch <- c
	defer func() { bridge.leavech <- c }()

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
		in := newBridgeInput(c, pg[:n])
		style.Println(bridge.name, "<-", ra,
			frame.Header[frame.ETH](in.pg))
		bridge.inputch <- in
	}
}

func Bridge(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	var flood []*tls.Conn
	lookup := make(map[uint64]*tls.Conn)

	bridge.name = Self().Name()

	for {
		select {
		case <-ctx.Done():
			flood = flood[:0]
			for k := range lookup {
				delete(lookup, k)
			}
			return
		case c := <-bridge.joinch:
			flood = append(flood, c)
		case c := <-bridge.leavech:
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
		case in := <-bridge.inputch:
			eth := frame.Header[frame.ETH](in.pg)
			if eth.ShouldLearn() {
				lookup[eth.SA()] = in.c
			}
			da := eth.DA()
			if c, ok := lookup[da]; ok {
				bridgeSend(ctx, c, in)
			} else {
				for _, c := range flood {
					if c != in.c {
						bridgeSend(ctx, c, in)
					}
				}
			}
			in.free()
		}
	}
}

func bridgeSend(ctx context.Context, c *tls.Conn, in *bridgeInput) {
	ra := remoteAddr(c)
	_, err := lv.NewEncoder(write.With(ctx, c)).Write(in.pg)
	if err != nil {
		style.Errorln(bridge.name, "->", ra, err)
	} else {
		style.Println(bridge.name, "->", ra,
			frame.Header[frame.ETH](in.pg))
	}
}

func newBridgeInput(c *tls.Conn, pg []byte) *bridgeInput {
	in := bridge.pool.Get().(*bridgeInput)
	in.c = c
	in.pg = pg
	return in
}

func (in *bridgeInput) free() {
	page.Free(in.pg)
	in.c = nil
	in.pg = in.pg[:0]
	bridge.pool.Put(in)
}
