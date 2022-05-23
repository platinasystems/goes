// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package lv

import (
	"encoding/binary"
	"io"
)

func NewEncoder(w io.Writer) Encode { return Encode{w} }

type Encode struct{ w io.Writer }

// A zero length value to break context.
func (enc Encode) Break() error {
	var b [2]byte
	_, err := enc.w.Write(b[:])
	return err
}

// A negative acknowledgment.
func (enc Encode) Nack(err error) error {
	var b [2]byte
	data := []byte(err.Error())
	n := len(data)
	if n > Max {
		return ErrTooLarge
	}
	binary.BigEndian.PutUint16(b[:], uint16(n)|Eflag)
	if _, err = enc.w.Write(b[:]); err == nil {
		_, err = enc.w.Write(data)
	}
	return err
}

// Write an encoded length followed by data to the encapsulated writer.
func (enc Encode) Write(data []byte) (t int, err error) {
	var b [2]byte
	n := len(data)
	if n == 0 {
		return
	}
	for n > 0 {
		if n > Max {
			n = Max
		}
		binary.BigEndian.PutUint16(b[:], uint16(n))
		if _, err = enc.w.Write(b[:]); err != nil {
			break
		}
		if _, err = enc.w.Write(data[:n]); err != nil {
			break
		}
		t += n
		data = data[n:]
		n = len(data)
	}
	return
}

func (enc Encode) WriteString(s string) (int, error) {
	return enc.Write([]byte(s))
}
