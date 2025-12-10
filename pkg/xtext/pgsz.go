// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xtext

import (
	"io"
	"os"
	"strconv"

	"golang.org/x/term"
)

const (
	DefaultPageWidth  = 80
	DefaultPageHeight = 24
)

type PageSize struct {
	Width,
	Height uint
}

// If writer is terminal, try [term.GetSize].
// If unsuccessful, try “COLUMNS” and “ROWS' environment variables.
// If still unavailable, return [DefaultPageWidth]  and [DefaultPageHeight].
func GetPageSize(w io.Writer) (pg PageSize) {
	if v, ok := w.(interface{ Fd() uintptr }); ok {
		if fd := int(v.Fd()); term.IsTerminal(fd) {
			if w, h, err := term.GetSize(fd); err == nil {
				pg.Width = uint(w)
				pg.Height = uint(h)
			}
		}
	}
	if pg.Width == 0 {
		pg.Width = DefaultPageWidth
		if s, ok := os.LookupEnv("COLUMNS"); ok {
			if u, err := strconv.ParseUint(s, 10, 32); err == nil {
				pg.Width = uint(u)
			}
		}
	}
	if pg.Height == 0 {
		pg.Height = DefaultPageHeight
		if s, ok := os.LookupEnv("ROWS"); ok {
			if u, err := strconv.ParseUint(s, 10, 32); err == nil {
				pg.Height = uint(u)
			}
		}
	}
	return
}
