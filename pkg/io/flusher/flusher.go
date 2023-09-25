// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package flusher

import (
	"io"

	"github.com/platinasystems/goes/v2/pkg/sync/chunk"
)

type ReadWriterTo interface {
	io.Reader
	io.WriterTo
}

var (
	alloc = chunk.New4K
	free  = chunk.Free4K
)

// The returned flush() reads the given Reader until EOF.
func New(r io.Reader) (wt ReadWriterTo, flush func()) {
	in := &input{r: r}
	wt = in
	flush = in.flush
	return
}

type input struct {
	r io.Reader
	e error
}

func (in *input) Read(b []byte) (int, error) {
	var n int
	if in.r == nil {
		return 0, io.EOF
	}
	if in.e != nil {
		return 0, in.e
	}
	n, in.e = in.r.Read(b)
	if n == 0 && in.e == nil {
		in.e = io.EOF
	}
	return n, in.e
}

func (in *input) WriteTo(w io.Writer) (int64, error) {
	var n int64
	if in.r == nil {
		return 0, io.EOF
	}
	if in.e != nil {
		return 0, in.e
	}
	buf := alloc()
	defer free(buf)
	for {
		var nr, nw int
		if nr, in.e = in.r.Read(buf); in.e != nil || nr == 0 {
			break
		}
		nw, in.e = w.Write(buf[:nr])
		if nw > 0 {
			n += int64(nw)
		}
		if in.e != nil || nw != nr {
			break
		}
	}
	in.r = nil
	return n, in.e
}

func (in *input) flush() {
	if in.r != nil && in.e == nil {
		buf := alloc()
		defer free(buf)
		for {
			n, err := in.r.Read(buf)
			if n == 0 || err != nil {
				break
			}
		}
	}
	in.r = nil
	in.e = nil
}
