// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package publisher

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
)

func New() (*Publisher, error) {
	addr, err := net.ResolveUnixAddr("unixgram", "@redis.pub")
	return &Publisher{
		addr: addr,
		buf:  new(bytes.Buffer),
	}, err
}

type Publisher struct {
	sync.Mutex
	addr *net.UnixAddr
	conn *net.UnixConn
	buf  *bytes.Buffer
}

func (p *Publisher) Close() error {
	var err error
	p.Lock()
	defer p.Unlock()
	if p.conn != nil {
		err = p.conn.Close()
		p.conn = nil
	}
	p.addr = nil
	return err
}

func (p *Publisher) Print(a ...interface{}) (int, error) {
	return p.flush(func(buf *bytes.Buffer) (int, error) {
		return fmt.Fprint(buf, a...)
	})
}

func (p *Publisher) Printf(format string, a ...interface{}) (int, error) {
	return p.flush(func(buf *bytes.Buffer) (int, error) {
		return fmt.Fprintf(buf, format, a...)
	})
}

func (p *Publisher) Write(b []byte) (int, error) {
	return p.flush(func(buf *bytes.Buffer) (int, error) {
		return buf.Write(b)
	})
}

func (p *Publisher) flush(fill func(*bytes.Buffer) (int, error)) (int, error) {
	p.Lock()
	defer p.Unlock()
	if p.addr == nil {
		return 0, io.EOF
	}
	if p.conn == nil {
		conn, err := net.DialUnix("unixgram", nil, p.addr)
		if err != nil {
			return 0, err
		}
		p.conn = conn
	}
	p.buf.Reset()
	n, err := fill(p.buf)
	if err == nil && p.buf.Len() > 0 {
		if n, err = p.conn.Write(p.buf.Bytes()); err != nil {
			p.conn.Close()
			p.conn = nil
		}
	}
	return n, err
}
