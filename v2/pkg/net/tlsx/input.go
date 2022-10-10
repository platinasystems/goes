// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"io"

	"github.com/platinasystems/goes/v2/pkg/os/page"
)

type input struct {
	r io.Reader
	e error
}

func newinput(r io.Reader) (io.Reader, func()) {
	in := &input{r: r}
	return in, in.flush
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

func (in *input) flush() {
	if in.r != nil && in.e == nil {
		pg := page.New()
		defer page.Free(pg)
		for {
			n, err := in.r.Read(pg)
			if n == 0 || err != nil {
				break
			}
		}
	}
	in.r = nil
	in.e = nil
}
