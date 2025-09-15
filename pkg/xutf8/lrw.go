// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xutf8

import (
	"io"
	"unicode/utf8"
)

// NewLastRuneWrapper wraps an [io.Writer] to record the last written rune.
//
//	lrw := NewLastRuneWrapper(w)
//	lrw.Write(...)
//	if lrw.LastWrittenRune() != '\n' {
//		w.Write([]byte("\n")
//	}
func NewLastRuneWrapper(w io.Writer) interface {
	io.Writer
	LastWrittenRune() rune
} {
	return &lrw{w: w}
}

type lrw struct {
	r rune
	w io.Writer
}

func (lrw *lrw) LastWrittenRune() rune {
	return lrw.r
}

func (lrw *lrw) Write(b []byte) (int, error) {
	n, err := lrw.w.Write(b)
	if err == nil {
		lrw.r, _ = utf8.DecodeLastRune(b)
	}
	return n, err
}
