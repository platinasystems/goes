// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import (
	"encoding/binary"
)

// [binary.BigEndian] encode “val” to end of “buf” and,
// if successful, return the expanded buffer up to it's capacity.
func Add(buf []byte, val any) ([]byte, error) {
	n, err := binary.Encode(buf[len(buf):cap(buf)], binary.BigEndian, val)
	if err == nil {
		buf = buf[:len(buf)+n]
	}
	return buf, err
}

// [binary.BigEndian] decode from beginning of “data” to “ptr” and,
// if successful, return remainder.
func Subtract(data []byte, ptr any) ([]byte, error) {
	n, err := binary.Decode(data, binary.BigEndian, ptr)
	if err == nil {
		data = data[n:]
	}
	return data, err
}
