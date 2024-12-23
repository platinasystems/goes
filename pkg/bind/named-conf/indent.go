// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import (
	"bytes"
	"io"
	"unicode/utf8"
)

type Indent struct {
	w    io.Writer
	i    int
	last rune
}

func NewIndent(w io.Writer) *Indent {
	iw := &Indent{w: w, i: 2}
	if cur, ok := w.(*Indent); ok {
		iw.i += cur.i
	}
	return iw
}

func (iw *Indent) Write(b []byte) (int, error) {
	const indentation = "                    "
	var n int
	nlsz := utf8.RuneLen('\n')
	for len(b) > 0 {
		if iw.last == '\n' {
			iw.w.Write([]byte(indentation[:iw.i]))
		}
		i := bytes.IndexRune(b, '\n')
		if i < 0 {
			wrote, err := iw.w.Write(b)
			if wrote > 0 {
				n += wrote
			}
			iw.last, _ = utf8.DecodeLastRune(b)
			return n, err
		}
		wrote, err := iw.w.Write(b[:i+nlsz])
		if wrote > 0 {
			n += wrote
		}
		if err != nil {
			return n, err
		}
		b = b[i+nlsz:]
		iw.last = '\n'
	}
	return n, nil

}
