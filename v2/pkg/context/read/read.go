// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package read

import (
	"context"
	"io"
)

// This is a contextual reader that checks the embedded context Err() and
// Done() before or after read.
func With(ctx context.Context, r io.Reader) io.Reader {
	return cr{ctx, r}
}

type cr struct {
	c context.Context
	r io.Reader
}

func (cr cr) Read(buf []byte) (int, error) {
	err := cr.c.Err()
	if err != nil {
		return 0, err
	}
	if cr.r == nil {
		return 0, io.EOF
	}
	select {
	case <-cr.c.Done():
		return 0, context.Canceled
	default:
	}
	n, err := cr.r.Read(buf)
	if err == nil {
		select {
		case <-cr.c.Done():
			err = context.Canceled
		default:
			err = cr.c.Err()
		}
	}
	return n, err
}
