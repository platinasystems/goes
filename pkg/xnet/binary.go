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

var (
	Uint16 = binary.BigEndian.Uint16
	Uint32 = binary.BigEndian.Uint32
	Uint64 = binary.BigEndian.Uint64
)

func Attach16[I Int | Uint](buf []byte, i I) []byte {
	return binary.BigEndian.AppendUint16(buf, uint16(i))
}

func Attach32[I Int | Uint](buf []byte, i I) []byte {
	return binary.BigEndian.AppendUint32(buf, uint32(i))
}

func Attach64[I Int | Uint](buf []byte, i I) []byte {
	return binary.BigEndian.AppendUint64(buf, uint64(i))
}

func Decode16[I Int | Uint](data []byte, p *I) {
	*p = I(Uint16(data))
}

func Decode32[I Int | Uint](data []byte, p *I) {
	*p = I(Uint32(data))
}

func Decode64[I Int | Uint](data []byte, p *I) {
	*p = I(Uint64(data))
}

// Decode and trim last 2-byte value from buffer.
func Detach16[I Int | Uint](buf []byte, p *I) []byte {
	Decode16(buf, p)
	return buf[:len(buf)-2]
}

// Decode and trim last 4-byte value from buffer.
func Detach32[I Int | Uint](buf []byte, p *I) []byte {
	Decode32(buf, p)
	return buf[:len(buf)-4]
}

// Decode and trim last 8-byte value from buffer.
func Detach64[I Int | Uint](buf []byte, p *I) []byte {
	Decode64(buf, p)
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

func Int16(buf []byte) int16 { return int16(Uint16(buf)) }
func Int32(buf []byte) int32 { return int32(Uint32(buf)) }
func Int64(buf []byte) int64 { return int64(Uint64(buf)) }

// Decode and trim first 2-byte value from buffer.
func Remove16[I Int | Uint](data []byte, p *I) []byte {
	Decode16(data, p)
	return data[2:]
}

// Decode and trim first 4-byte value from buffer.
func Remove32[I Int | Uint](data []byte, p *I) []byte {
	Decode32(data, p)
	return data[4:]
}

// Decode and trim first 8-byte value from buffer.
func Remove64[I Int | Uint](data []byte, p *I) []byte {
	Decode64(data, p)
	return data[8:]
}

// [binary.BigEndian] [binary.Append] to end of bytes.
func Attach(buf []byte, val any) ([]byte, error) {
	return binary.Append(buf, binary.BigEndian, val)
}

// [binary.BigEndian] [binary.Encode] to beginning of bytes.
func Encode(buf []byte, val any) (int, error) {
	return binary.Encode(buf, binary.BigEndian, val)
}

// [binary.BigEndian] [binary.Decode] from beginning of bytes and
// if successful, return remainder.
func Remove(data []byte, ptr any) ([]byte, error) {
	n, err := binary.Decode(data, binary.BigEndian, ptr)
	if err == nil {
		data = data[n:]
	}
	return data, err
}
