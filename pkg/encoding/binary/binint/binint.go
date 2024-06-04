// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binint

import (
	"encoding/binary"
	"unsafe"
)

type Integer interface {
	Signed | Unsigned
}

type Signed interface {
	~int | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

var Endians = map[string]interface {
	binary.ByteOrder
	binary.AppendByteOrder
}{
	"big":    binary.BigEndian,
	"little": binary.LittleEndian,
	"native": binary.NativeEndian,
}

func Append[I Integer](abo binary.AppendByteOrder, data []byte, v I) []byte {
	switch unsafe.Sizeof(v) {
	case 2:
		return abo.AppendUint16(data, uint16(v))
	case 4:
		return abo.AppendUint32(data, uint32(v))
	case 8:
		return abo.AppendUint64(data, uint64(v))
	}
	return data
}

func AppendBig[I Integer](data []byte, v I) []byte {
	return Append[I](binary.BigEndian, data, v)
}

func AppendLittle[I Integer](data []byte, v I) []byte {
	return Append[I](binary.LittleEndian, data, v)
}

func AppendNative[I Integer](data []byte, v I) []byte {
	return Append[I](binary.NativeEndian, data, v)
}

func Bite[I ~int8 | ~uint8](data []byte, p *I) []byte {
	*p = I(data[0])
	return data[1:]
}

func Pull[I Integer](bo binary.ByteOrder, data []byte, p *I) []byte {
	switch unsafe.Sizeof(*p) {
	case 2:
		*p = I(bo.Uint16(data))
		data = data[2:]
	case 4:
		*p = I(bo.Uint32(data))
		data = data[4:]
	case 8:
		*p = I(bo.Uint64(data))
		data = data[8:]
	}
	return data
}

func PullBig[I Integer](data []byte, p *I) []byte {
	return Pull[I](binary.BigEndian, data, p)
}

func PullLittle[I Integer](data []byte, p *I) []byte {
	return Pull[I](binary.LittleEndian, data, p)
}

func PullNative[I Integer](data []byte, p *I) []byte {
	return Pull[I](binary.NativeEndian, data, p)
}
