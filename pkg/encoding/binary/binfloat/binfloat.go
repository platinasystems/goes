// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binfloat

import (
	"math"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

type Endian = binint.Endian

type Float interface {
	~float32 | ~float64
}

var Big = binint.Big
var Little = binint.Little
var Native = binint.Native

func Append[F Float](endian Endian, data []byte, f F) []byte {
	switch unsafe.Sizeof(f) {
	case 4:
		data = binint.Append(endian, data, math.Float32bits(float32(f)))
	case 8:
		data = binint.Append(endian, data, math.Float64bits(float64(f)))
	}
	return data
}

func AppendBig[F Float](data []byte, f F) []byte {
	return Append(Big, data, f)
}

func AppendLittle[F Float](data []byte, f F) []byte {
	return Append(Little, data, f)
}

func AppendNative[F Float](data []byte, f F) []byte {
	return Append(Native, data, f)
}

func New[F Float](endian Endian, f F) (data []byte) {
	switch unsafe.Sizeof(f) {
	case 4:
		data = binint.New(endian, math.Float32bits(float32(f)))
	case 8:
		data = binint.New(endian, math.Float64bits(float64(f)))
	}
	return
}

func NewBig[F Float](f F) []byte {
	return New(Big, f)
}

func NewLittle[F Float](f F) []byte {
	return New(Little, f)
}

func NewNative[F Float](f F) []byte {
	return New(Native, f)
}

func Pull[F Float](endian Endian, data []byte) (f F, rem []byte) {
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

func PullBig[F Float](data []byte) (f F, rem []byte) {
	return Pull[F](Big, data)
}

func PullLittle[F Float](data []byte) (f F, rem []byte) {
	return Pull[F](Little, data)
}

func PullNative[F Float](data []byte) (f F, rem []byte) {
	return Pull[F](Native, data)
}
