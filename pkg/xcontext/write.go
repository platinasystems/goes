// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xcontext

import (
	. "context"
	"io"
)

// This a contextual writer that checks the embedded context Err() before or
// after write.
type ContextualWriter struct {
	Context
	w io.Writer
}

func WithWriter(ctx Context, w io.Writer) ContextualWriter {
	return ContextualWriter{ctx, w}
}

func (cw ContextualWriter) Write(data []byte) (int, error) {
	err := cw.Err()
	if err != nil {
		return 0, err
	}
	n, err := cw.w.Write(data)
	if err == nil {
		err = cw.Err()
	}
	return n, err
}
