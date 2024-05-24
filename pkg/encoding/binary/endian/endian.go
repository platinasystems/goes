// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package endian

import (
	"encoding/binary"
	"math"
	"unsafe"

	"golang.org/x/exp/constraints"
)

var Big = binary.BigEndian
var Little = binary.LittleEndian
var Native = binary.NativeEndian

type Endian interface {
	binary.ByteOrder
	binary.AppendByteOrder
}

func AppendBigFloat[F constraints.Float](data []byte, f F) []byte {
	return AppendFloat(Big, data, f)
}

func AppendBigInteger[I constraints.Integer](data []byte, i I) []byte {
	return AppendInteger(Big, data, i)
}

func AppendFloat[F constraints.Float](
	endian Endian, data []byte, f F,
) []byte {
	switch unsafe.Sizeof(f) {
	case 4:
		return AppendInteger(endian, data, math.Float32bits(float32(f)))
	case 8:
		return AppendInteger(endian, data, math.Float64bits(float64(f)))
	}
	return data
}

func AppendInteger[I constraints.Integer](
	endian Endian, data []byte, i I,
) []byte {
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

func AppendLittleFloat[F constraints.Float](data []byte, f F) []byte {
	return AppendFloat(Little, data, f)
}

func AppendLittleInteger[I constraints.Integer](data []byte, i I) []byte {
	return AppendInteger(Little, data, i)
}

func AppendNativeFloat[F constraints.Float](data []byte, f F) []byte {
	return AppendFloat(Native, data, f)
}

func AppendNativeInteger[I constraints.Integer](data []byte, i I) []byte {
	return AppendInteger(Native, data, i)
}

func NewBigInteger[I constraints.Integer](i I) []byte {
	return NewInteger(Big, i)
}

func NewBigFloat[F constraints.Float](f F) []byte {
	return NewFloat(Big, f)
}

func NewFloat[F constraints.Float](endian Endian, f F) (data []byte) {
	switch unsafe.Sizeof(f) {
	case 4:
		data = NewInteger(endian, math.Float32bits(float32(f)))
	case 8:
		data = NewInteger(endian, math.Float64bits(float64(f)))
	}
	return
}

func NewInteger[I constraints.Integer](endian Endian, i I) (data []byte) {
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

func NewLittleFloat[F constraints.Float](f F) []byte {
	return NewFloat(Little, f)
}

func NewLittleInteger[I constraints.Integer](i I) []byte {
	return NewInteger(Little, i)
}

func NewNativeFloat[F constraints.Float](f F) []byte {
	return NewFloat(Native, f)
}

func NewNativeInteger[I constraints.Integer](i I) []byte {
	return NewInteger(Native, i)
}

func PullBigFloat[F constraints.Float](data []byte) (f F, rem []byte) {
	return PullFloat[F](Big, data)
}

func PullFloat[F constraints.Float](
	endian Endian, data []byte,
) (f F, rem []byte) {
	switch unsafe.Sizeof(f) {
	case 4:
		f = F(math.Float32frombits(endian.Uint32(data)))
		rem = data[4:]
	case 8:
		f = F(math.Float64frombits(endian.Uint64(data)))
		rem = data[8:]
	}
	return
}

func PullBigInteger[I constraints.Integer](data []byte) (i I, rem []byte) {
	return PullInteger[I](Big, data)
}

func PullInteger[I constraints.Integer](
	endian Endian, data []byte,
) (i I, rem []byte) {
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

func PullLittleFloat[F constraints.Float](data []byte) (f F, rem []byte) {
	return PullFloat[F](Little, data)
}

func PullLittleInteger[I constraints.Integer](data []byte) (i I, rem []byte) {
	return PullInteger[I](Little, data)
}

func PullNativeFloat[F constraints.Float](data []byte) (f F, rem []byte) {
	return PullFloat[F](Native, data)
}

func PullNativeInteger[I constraints.Integer](data []byte) (i I, rem []byte) {
	return PullInteger[I](Native, data)
}
