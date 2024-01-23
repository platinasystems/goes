// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ensure

import (
	"bytes"
	"io"
)

// Newline returns a wrapper that, if necessary, writes a trailing newline
// when closed.
func Newline(w io.Writer) io.WriteCloser {
	return &nw{w: w}
}

type nw struct {
	w io.Writer
	b bytes.Buffer
}

var nl = []byte{'\n'}

func (p *nw) Write(b []byte) (int, error) {
	p.b.Reset()
	p.b.Write(b)
	return p.w.Write(b)
}

func (p *nw) Close() error {
	if p == nil || p.w == nil {
		return EINVAL
	}
	if !bytes.HasSuffix(p.b.Bytes(), nl) {
		p.w.Write(nl)
	}
	p.b.Reset()
	p.w = nil
	return nil
}
