// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package lv

import (
	"encoding/binary"
	"io"
)

type Decoding struct{ r io.Reader }

// This returns a wrapper that reads an encoded length and the following data
// from the encapsulated reader.  If the encoded length has Ebit, the data is
// returned as a Nack error.
func NewDecoder(r io.Reader) Decoding { return Decoding{r} }

func (dec Decoding) Read(b []byte) (int, error) {
	if n := len(b); n < 2 {
		return 0, ErrTooSmall
	}
	n, err := io.ReadFull(dec.r, b[:2])
	if err != nil {
		return 0, err
	}
	u := binary.BigEndian.Uint16(b[:2])
	e := u & Eflag
	n = int(u & Efilter)
	switch {
	case n == 0:
		return 0, nil
	case n > Max:
		return 0, ErrTooLarge
	case n > len(b):
		return 0, ErrTooSmall
	}
	_, err = io.ReadFull(dec.r, b[:n])
	if err != nil {
		n = 0
	} else if e == Eflag {
		s := string(b[:n])
		err = NewNack(s)
		n = 0
	}
	return n, err
}
