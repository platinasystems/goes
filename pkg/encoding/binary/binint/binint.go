// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binint

import (
	"encoding/binary"
	"unsafe"
)

type Endian interface {
	binary.ByteOrder
	binary.AppendByteOrder
}

type Integer interface {
	Signed | Unsigned
}

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

var Big = binary.BigEndian
var Little = binary.LittleEndian
var Native = binary.NativeEndian

var Endians = map[string]Endian{
	"big":    Big,
	"little": Little,
	"native": Native,
}

func Append[I Integer](endian Endian, data []byte, i I) []byte {
	switch unsafe.Sizeof(i) {
	case 1:
		data = append(data, byte(i))
	case 2:
		data = endian.AppendUint16(data, uint16(i))
	case 4:
		data = endian.AppendUint32(data, uint32(i))
	case 8:
		data = endian.AppendUint64(data, uint64(i))
	}
	return data
}

func AppendBig[I Integer](data []byte, i I) []byte {
	return Append(Big, data, i)
}

func AppendLittle[I Integer](data []byte, i I) []byte {
	return Append(Little, data, i)
}

func AppendNative[I Integer](data []byte, i I) []byte {
	return Append(Native, data, i)
}

func New[I Integer](endian Endian, i I) (data []byte) {
	switch unsafe.Sizeof(i) {
	case 1:
		data = make([]byte, 1, 1)
		data[0] = byte(i)
	case 2:
		data = make([]byte, 2, 2)
		endian.PutUint16(data, uint16(i))
	case 4:
		data = make([]byte, 4, 4)
		endian.PutUint32(data, uint32(i))
	case 8:
		data = make([]byte, 8, 8)
		endian.PutUint64(data, uint64(i))
	}
	return
}

func NewBig[I Integer](i I) []byte {
	return New(Big, i)
}

func NewLittle[I Integer](i I) []byte {
	return New(Little, i)
}

func NewNative[I Integer](i I) []byte {
	return New(Native, i)
}

func Pull[I Integer](endian Endian, data []byte) (i I, rem []byte) {
	switch unsafe.Sizeof(i) {
	case 1:
		i, rem = I(data[0]), data[1:]
	case 2:
		i, rem = I(endian.Uint16(data)), data[2:]
	case 4:
		i, rem = I(endian.Uint32(data)), data[4:]
	case 8:
		i, rem = I(endian.Uint64(data)), data[8:]
	}
	return
}

func PullBig[I Integer](data []byte) (i I, rem []byte) {
	return Pull[I](Big, data)
}

func PullLittle[I Integer](data []byte) (i I, rem []byte) {
	return Pull[I](Little, data)
}

func PullNative[I Integer](data []byte) (i I, rem []byte) {
	return Pull[I](Native, data)
}
