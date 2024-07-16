// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package lv

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Encoding struct{ w io.Writer }

func NewEncoder(w io.Writer) Encoding { return Encoding{w} }

// Implied arg type encoding:
//
//	  []any	Recurse
//
//	  []string
//		Iterate
//
//	  nil	Break
//
//	  error	Nack
//
//	  []byte, string
//		Write
//
// Otherwise, [fmt.Fprint] anything else.
func (enc Encoding) Encode(args ...any) (n int, err error) {
	var brk [2]byte
	for _, arg := range args {
		var i int
		switch t := arg.(type) {
		case nil:
			// A zero length value to break context.
			_, err = enc.w.Write(brk[:])
		case error:
			err = enc.nack(t)
			i = len(t.Error())
		case []byte:
			if len(t) > 0 {
				i, err = enc.Write(t)
			} else {
				_, err = enc.w.Write(brk[:])
			}
		case string:
			i, err = enc.Write([]byte(t))
		case []any:
			i, err = enc.Encode(t...)
		case []string:
			for _, s := range t {
				if i, err = enc.Encode(s); err != nil {
					return
				}
				n += i
				i = 0
			}
		default:
			i, err = fmt.Fprint(enc, arg)
		}
		if err != nil {
			return
		}
		n += i
	}
	return
}

// Write an encoded length followed by data to the encapsulated writer.
func (enc Encoding) Write(data []byte) (t int, err error) {
	var nb [2]byte
	n := len(data)
	if n == 0 {
		return
	}
	for n > 0 {
		if n > Max {
			n = Max
		}
		binary.BigEndian.PutUint16(nb[:], uint16(n))
		if _, err = enc.w.Write(nb[:]); err != nil {
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

// A negative acknowledgment.
func (enc Encoding) nack(err error) error {
	var nb [2]byte
	data := []byte(err.Error())
	n := len(data)
	if n > Max {
		return fmt.Errorf("nack length: %d: too large", n)
	}
	binary.BigEndian.PutUint16(nb[:], uint16(n)|Eflag)
	if _, err = enc.w.Write(nb[:]); err == nil {
		_, err = enc.w.Write(data)
	}
	return err
}
