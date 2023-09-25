// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package read

import (
	"context"
	"io"
)

// This is a contextual reader that checks the embedded context Err() before or
// after read.
type ContextualReader struct {
	context.Context
	r io.Reader
}

func With(ctx context.Context, r io.Reader) ContextualReader {
	return ContextualReader{ctx, r}
}

func (cr ContextualReader) Read(buf []byte) (int, error) {
	err := cr.Err()
	if err != nil {
		return 0, err
	}
	if cr.r == nil {
		return 0, io.EOF
	}
	n, err := cr.r.Read(buf)
	if err == nil {
		err = cr.Err()
	}
	return n, err
}
