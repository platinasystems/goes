// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package publisher

import (
	"bytes"
	"fmt"
	"io"
	"net"

	"github.com/platinasystems/go/internal/accumulate"
	"github.com/platinasystems/go/internal/atsock"
)

const Name = "redis.pub"

func AtSock() string { return atsock.Name(Name) }

func New() (*Publisher, error) {
	a, err := net.ResolveUnixAddr("unixgram", AtSock())
	if err != nil {
		return nil, err
	}
	sock, err := net.DialUnix("unixgram", nil, a)
	return &Publisher{
		Accumulator: accumulate.Accumulator{
			ReaderOrWriter: sock,
		},
		buf: new(bytes.Buffer),
	}, nil
}

type Publisher struct {
	accumulate.Accumulator
	buf *bytes.Buffer
}

func (p *Publisher) Close() error {
	return p.Accumulator.ReaderOrWriter.(io.Closer).Close()
}

func (p *Publisher) Print(a ...interface{}) (int, error) {
	p.buf.Reset()
	n, err := fmt.Fprint(p.buf, a...)
	if err == nil {
		p.Write(p.buf.Bytes())
	}
	return n, err
}

func (p *Publisher) Printf(format string, a ...interface{}) (int, error) {
	p.buf.Reset()
	n, err := fmt.Fprintf(p.buf, format, a...)
	if err == nil {
		p.Write(p.buf.Bytes())
	}
	return n, err
}
