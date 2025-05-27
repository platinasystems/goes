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

func Attach16[I Int | Uint](buf []byte, i I) []byte {
	return binary.BigEndian.AppendUint16(buf, uint16(i))
}

func Attach32[I Int | Uint](buf []byte, i I) []byte {
	return binary.BigEndian.AppendUint32(buf, uint32(i))
}

func Attach64[I Int | Uint](buf []byte, i I) []byte {
	return binary.BigEndian.AppendUint64(buf, uint64(i))
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

func Detach16[I Int | Uint](buf []byte, p *I) []byte {
	*p = I(binary.BigEndian.Uint16(buf[len(buf)-2:]))
	return buf[:len(buf)-2]
}

func Detach32[I Int | Uint](buf []byte, p *I) []byte {
	*p = I(binary.BigEndian.Uint32(buf[len(buf)-4:]))
	return buf[:len(buf)-4]
}

func Detach64[I Int | Uint](buf []byte, p *I) []byte {
	*p = I(binary.BigEndian.Uint64(buf[len(buf)-8:]))
	return buf[:len(buf)-8]
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

func Remove16[I Int | Uint](data []byte, p *I) []byte {
	*p = I(binary.BigEndian.Uint16(data))
	return data[2:]
}

func Remove32[I Int | Uint](data []byte, p *I) []byte {
	*p = I(binary.BigEndian.Uint32(data))
	return data[4:]
}

func Remove64[I Int | Uint](data []byte, p *I) []byte {
	*p = I(binary.BigEndian.Uint64(data))
	return data[8:]
}

// [binary.BigEndian] append to end of bytes.
func Attach(buf []byte, val any) ([]byte, error) {
	return binary.Append(buf, binary.BigEndian, val)
}

// [binary.BigEndian] encode to beginning of bytes.
func Encode(buf []byte, val any) (int, error) {
	return binary.Encode(buf, binary.BigEndian, val)
}

// [binary.BigEndian] decode from beginning of bytes and
// if successful, return remainder.
func Remove(data []byte, ptr any) ([]byte, error) {
	n, err := binary.Decode(data, binary.BigEndian, ptr)
	if err == nil {
		data = data[n:]
	}
	return data, err
}
