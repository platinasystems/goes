// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import "encoding/binary"

type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Uint interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func Decode16[I Int | Uint](data []byte) I {
	return I(binary.BigEndian.Uint16(data))
}

func Decode32[I Int | Uint](data []byte) I {
	return I(binary.BigEndian.Uint32(data))
}

func Decode64[I Int | Uint](data []byte) I {
	return I(binary.BigEndian.Uint64(data))
}

func Encode16[I Int | Uint](buf []byte, i I) {
	binary.BigEndian.PutUint16(buf, uint16(i))
}

func Encode32[I Int | Uint](buf []byte, i I) {
	binary.BigEndian.PutUint32(buf, uint32(i))
}

func Encode64[I Int | Uint](buf []byte, i I) {
	binary.BigEndian.PutUint64(buf, uint64(i))
}

// [binary.BigEndian] encode “val” to end of “buf” and,
// if successful, return the expanded buffer up to it's capacity.
func Attach(buf []byte, val any) ([]byte, error) {
	n, err := binary.Encode(buf[len(buf):cap(buf)], binary.BigEndian, val)
	if err == nil {
		buf = buf[:len(buf)+n]
	}
	return buf, err
}

// [binary.BigEndian] encode “val” to beginning of “buf”.
func Encode(buf []byte, val any) (int, error) {
	return binary.Encode(buf, binary.BigEndian, val)
}

// [binary.BigEndian] decode from beginning of “data” to “ptr” and,
// if successful, return remainder.
func Remove(data []byte, ptr any) ([]byte, error) {
	n, err := binary.Decode(data, binary.BigEndian, ptr)
	if err == nil {
		data = data[n:]
	}
	return data, err
}
