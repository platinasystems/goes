// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bridge

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/frame"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/exchange/bridge/leasing"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/remote"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

var (
	ErrTooShort    = errors.New("len(frame) < 14")
	ErrUnavailable = errors.New("disabled bridge")
)

type T struct {
	Enabled bool
	Leasing leasing.T
	name    string
	inputch chan *input
	joinch  chan *tls.Conn
	leavech chan *tls.Conn
}

func (t *T) Configure(args []string) ([]string, error) {
	if len(args) > 0 {
		switch args[0] {
		case "-h":
			return args, flag.ErrHelp
		case "leasing":
			var err error
			args, err = t.Leasing.Configure(args[1:])
			if err != nil {
				return args, err
			}
		}
	}
	t.name = certs.Self.Name()
	t.inputch = make(chan *input, 16)
	t.joinch = make(chan *tls.Conn, 4)
	t.leavech = make(chan *tls.Conn, 4)
	t.Enabled = true
	return args, nil
}

func (t *T) Join(ctx context.Context, c *tls.Conn, args []string) {
	var tenant string

	t.joinch <- c
	defer func() { t.leavech <- c }()

	ra := remote.Addr(c)
	sub := dns0(c)
	dec := lv.NewDecoder(poll.With(ctx, c))
	enc := lv.NewEncoder(write.With(ctx, c))

	if !t.Enabled {
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
	if !t.Leasing.Enabled {
		enc.Encode("OK", nil)
	} else if len(args) == 0 {
		enc.Encode(t.Leasing.Lease(tenant).String(), nil)
	} else if err := t.Leasing.Occupy(tenant, args[0]); err == nil {
		enc.Encode("OK", nil)
	} else {
		enc.Encode(err)
		return
	}
	style.Println("OK")
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
		style.Println(t.name, "<-", sub, frame.NewEth(in.pg))
		t.inputch <- in
	}
}

func (t *T) Routine(ctx context.Context, wg *sync.WaitGroup) {
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
		case c := <-t.joinch:
			flood = append(flood, c)
		case c := <-t.leavech:
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
		case in := <-t.inputch:
			eth := frame.NewEth(in.pg)
			if eth.ShouldLearn() {
				lookup[eth.SA()] = in.c
			}
			da := eth.DA()
			if c, ok := lookup[da]; ok {
				t.send(ctx, c, in)
			} else {
				for _, c := range flood {
					if c != in.c {
						t.send(ctx, c, in)
					}
				}
			}
			in.free()
		}
	}
}

func (t *T) send(ctx context.Context, c *tls.Conn, in *input) {
	_, err := lv.NewEncoder(write.With(ctx, c)).Write(in.pg)
	if err != nil {
		style.Errorln(t.name, "->", dns0(c), err)
	} else {
		style.Println(t.name, "->", dns0(c), frame.NewEth(in.pg))
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
