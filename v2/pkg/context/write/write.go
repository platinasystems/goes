// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package write

import (
	"context"
	"io"
)

// This a contextual writer that checks the embedded context Err() and
// Done() before or after write.
func With(ctx context.Context, w io.Writer) io.Writer {
	return cw{ctx, w}
}

type cw struct {
	c context.Context
	w io.Writer
}

func (cw cw) Write(data []byte) (int, error) {
	err := cw.c.Err()
	if err != nil {
		return 0, err
	}
	select {
	case <-cw.c.Done():
		return 0, context.Canceled
	default:
	}
	n, err := cw.w.Write(data)
	if err == nil {
		select {
		case <-cw.c.Done():
			err = context.Canceled
		default:
			err = cw.c.Err()
		}
	}
	return n, err
}
