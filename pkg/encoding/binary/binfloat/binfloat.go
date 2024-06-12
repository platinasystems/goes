// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binfloat

import (
	"encoding/binary"
	"io"
	"math"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

type Floater interface{ ~float32 | ~float64 }

type Pointer[F Floater] struct {
	binary.ByteOrder
	P *F
}

func BigEndianPointer[F Floater](p *F) Pointer[F] {
	return Pointer[F]{binary.BigEndian, p}
}

func ByteOrderPointer[F Floater](bo binary.ByteOrder, p *F) Pointer[F] {
	return Pointer[F]{bo, p}
}

func LittleEndianPointer[F Floater](p *F) Pointer[F] {
	return Pointer[F]{binary.LittleEndian, p}
}

func NativeEndianPointer[F Floater](p *F) Pointer[F] {
	return Pointer[F]{binary.NativeEndian, p}
}

func (p Pointer[F]) ReadFrom(r io.Reader) (int64, error) {
	var n int64
	err := binint.ErrInvalid
	switch unsafe.Sizeof(*p.P) {
	case 4:
		var u uint32
		n, err = binint.ByteOrderPointer(p.ByteOrder, &u).ReadFrom(r)
		*p.P = F(math.Float32frombits(u))
	case 8:
		var u uint64
		n, err = binint.ByteOrderPointer(p.ByteOrder, &u).ReadFrom(r)
		*p.P = F(math.Float64frombits(u))
	}
	return n, err
}

type Value[F Floater] struct {
	binary.ByteOrder
	V F
}

func BigEndianValue[F Floater](v F) Value[F] {
	return Value[F]{binary.BigEndian, v}
}

func ByteOrderValue[F Floater](bo binary.ByteOrder, v F) Value[F] {
	return Value[F]{bo, v}
}

func LittleEndianValue[F Floater](v F) Value[F] {
	return Value[F]{binary.LittleEndian, v}
}

func NativeEndianValue[F Floater](v F) Value[F] {
	return Value[F]{binary.NativeEndian, v}
}

func (v Value[F]) WriteTo(w io.Writer) (int64, error) {
	switch unsafe.Sizeof(v.V) {
	case 4:
		u := math.Float32bits(float32(v.V))
		return binint.ByteOrderValue(v.ByteOrder, u).WriteTo(w)
	case 8:
		u := math.Float64bits(float64(v.V))
		return binint.ByteOrderValue(v.ByteOrder, u).WriteTo(w)
	}
	return 0, binint.ErrInvalid
}
