// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binfloat

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

type Floater interface{ ~float32 | ~float64 }

func Append[F Floater](apo binary.AppendByteOrder, data []byte, v F) []byte {
	switch unsafe.Sizeof(v) {
	case 4:
		data = binint.Append(apo, data, math.Float32bits(float32(v)))
	case 8:
		data = binint.Append(apo, data, math.Float64bits(float64(v)))
	}
	return data
}

func AppendBig[F Floater](data []byte, v F) []byte {
	return Append(binary.BigEndian, data, v)
}

func AppendLittle[F Floater](data []byte, v F) []byte {
	return Append(binary.LittleEndian, data, v)
}

func AppendNative[F Floater](data []byte, v F) []byte {
	return Append(binary.NativeEndian, data, v)
}

func Pull[F Floater](bo binary.ByteOrder, data []byte, p *F) []byte {
	switch unsafe.Sizeof(*p) {
	case 4:
		var u uint32
		data = binint.Pull(bo, data, &u)
		*p = F(math.Float32frombits(u))
	case 8:
		var u uint64
		data = binint.Pull(bo, data, &u)
		*p = F(math.Float64frombits(u))
	}
	return data
}

func PullBig[F Floater](data []byte, p *F) []byte {
	return Pull[F](binary.BigEndian, data, p)
}

func PullLittle[F Floater](data []byte, p *F) []byte {
	return Pull[F](binary.LittleEndian, data, p)
}

func PullNative[F Floater](data []byte, p *F) []byte {
	return Pull[F](binary.NativeEndian, data, p)
}
