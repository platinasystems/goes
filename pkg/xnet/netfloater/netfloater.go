// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package netfloater provides an [io.ReaderFrom] and [io.WriterTo] for float
// values from/to network.
package netfloater

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"unsafe"
)

var ErrInvalid = errors.New("invalid")

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
	switch unsafe.Sizeof(*p.P) {
	case 4:
		var b [4]byte
		n, err := r.Read(b[:])
		if err != nil {
			return 0, err
		}
		u := p.ByteOrder.Uint32(b[:])
		*p.P = F(math.Float32frombits(u))
		return int64(n), nil
	case 8:
		var b [8]byte
		n, err := r.Read(b[:])
		if err != nil {
			return 0, err
		}
		u := p.ByteOrder.Uint64(b[:])
		*p.P = F(math.Float64frombits(u))
		return int64(n), nil
	}
	return 0, ErrInvalid
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
		var b [4]byte
		u := math.Float32bits(float32(v.V))
		v.ByteOrder.PutUint32(b[:], u)
		n, err := w.Write(b[:])
		return int64(n), err
	case 8:
		var b [8]byte
		u := math.Float64bits(float64(v.V))
		v.ByteOrder.PutUint64(b[:], u)
		n, err := w.Write(b[:])
		return int64(n), err
	}
	return 0, ErrInvalid
}
