// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package lv

import (
	"encoding/binary"
	"io"
)

// This returns a wrapper that reads an encoded length and the following data
// from the encapsulated reader.  If the encoded length has Ebit, the data is
// returned as a Nack error.
func NewDecoder(r io.Reader) io.Reader { return decode{r} }

type decode struct{ r io.Reader }

func (dec decode) Read(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, ErrTooSmall
	}
	_, err := io.ReadFull(dec.r, b[:2])
	if err != nil {
		return 0, err
	}
	u := binary.BigEndian.Uint16(b[:2])
	e := u & Eflag
	n := int(u & Efilter)
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
		err = nack(b[:n])
		n = 0
	}
	return n, err
}
