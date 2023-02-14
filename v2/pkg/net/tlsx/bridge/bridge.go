// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bridge

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/frame"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/lease"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/remote"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

type input struct {
	c  *tls.Conn
	pg []byte
}

var (
	ErrTooShort    = errors.New("too short")
	ErrUnavailable = errors.New("unavailable")
)

var (
	name    string
	inputch = make(chan *input, 16)
	joinch  = make(chan *tls.Conn, 4)
	leavech = make(chan *tls.Conn, 4)
	pool    = sync.Pool{New: func() any { return new(input) }}
)

func Join(ctx context.Context, c *tls.Conn, args []string) {
	var tenant string

	ra := remote.Addr(c)
	dec := lv.NewDecoder(poll.With(ctx, c))
	enc := lv.NewEncoder(write.With(ctx, c))

	if len(name) == 0 {
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
		if ski == certs.Self.SKI() {
			tenant += fmt.Sprint("@", ra)
		}
	}
	if p, err := lease.Contract(tenant, args...); err != nil {
		enc.Encode(err)
		return
	} else {
		enc.Encode(p.String(), nil)
	}

	joinch <- c
	defer func() { leavech <- c }()

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
		style.Println(name, "<-", ra, frame.NewEth(in.pg))
		inputch <- in
	}
}

func Routine(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	var flood []*tls.Conn
	lookup := make(map[uint64]*tls.Conn)

	name = certs.Self.Name()

	for {
		select {
		case <-ctx.Done():
			flood = flood[:0]
			for k := range lookup {
				delete(lookup, k)
			}
			return
		case c := <-joinch:
			flood = append(flood, c)
		case c := <-leavech:
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
		case in := <-inputch:
			eth := frame.NewEth(in.pg)
			if eth.ShouldLearn() {
				lookup[eth.SA()] = in.c
			}
			da := eth.DA()
			if c, ok := lookup[da]; ok {
				send(ctx, c, in)
			} else {
				for _, c := range flood {
					if c != in.c {
						send(ctx, c, in)
					}
				}
			}
			in.free()
		}
	}
}

func send(ctx context.Context, c *tls.Conn, in *input) {
	ra := remote.Addr(c)
	_, err := lv.NewEncoder(write.With(ctx, c)).Write(in.pg)
	if err != nil {
		style.Errorln(name, "->", ra, err)
	} else {
		style.Println(name, "->", ra, frame.NewEth(in.pg))
	}
}

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
