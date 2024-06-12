// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binint

import (
	"encoding/binary"
	"errors"
	"io"
	"unsafe"
)

var ErrInvalid = errors.New("invalid")

type Byte interface {
	~int8 | ~uint8
}

type Word interface {
	~uint | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~int | ~int16 | ~int32 | ~int64
}

type Pointer[I Byte | Word] struct {
	binary.ByteOrder
	P *I
}

func BigEndianPointer[W Word](p *W) Pointer[W] {
	return Pointer[W]{binary.BigEndian, p}
}

func BytePointer[B Byte](p *B) Pointer[B] {
	return Pointer[B]{P: p}
}

func ByteOrderPointer[W Word](bo binary.ByteOrder, p *W) Pointer[W] {
	return Pointer[W]{bo, p}
}

func LittleEndianPointer[W Word](p *W) Pointer[W] {
	return Pointer[W]{binary.LittleEndian, p}
}

func NativeEndianPointer[W Word](p *W) Pointer[W] {
	return Pointer[W]{binary.NativeEndian, p}
}

func (p Pointer[I]) ReadFrom(r io.Reader) (int64, error) {
	var (
		a [8]byte
		n int
	)
	err := ErrInvalid
	switch unsafe.Sizeof(*p.P) {
	case 1:
		n, err = r.Read(a[:1])
		*p.P = I(a[0])
	case 2:
		n, err = r.Read(a[:2])
		*p.P = I(p.Uint16(a[:2]))
	case 4:
		n, err = r.Read(a[:4])
		*p.P = I(p.Uint32(a[:4]))
	case 8:
		n, err = r.Read(a[:8])
		*p.P = I(p.Uint64(a[:8]))
	}
	return int64(n), err
}

type Value[I Byte | Word] struct {
	binary.ByteOrder
	V I
}

func BigEndianValue[W Word](v W) Value[W] {
	return Value[W]{binary.BigEndian, v}
}

func ByteValue[B Byte](v B) Value[B] {
	return Value[B]{V: v}
}

func ByteOrderValue[W Word](bo binary.ByteOrder, v W) Value[W] {
	return Value[W]{bo, v}
}

func LittleEndianValue[W Word](v W) Value[W] {
	return Value[W]{binary.LittleEndian, v}
}

func NativeEndianValue[W Word](v W) Value[W] {
	return Value[W]{binary.NativeEndian, v}
}

func (v Value[I]) WriteTo(w io.Writer) (int64, error) {
	var (
		a [8]byte
		n int
	)
	err := ErrInvalid
	switch unsafe.Sizeof(v.V) {
	case 1:
		a[0] = byte(v.V)
		n, err = w.Write(a[:1])
	case 2:
		v.PutUint16(a[:2], uint16(v.V))
		n, err = w.Write(a[:2])
	case 4:
		v.PutUint32(a[:4], uint32(v.V))
		n, err = w.Write(a[:4])
	case 8:
		v.PutUint64(a[:8], uint64(v.V))
		n, err = w.Write(a[:8])
	}
	return int64(n), err
}
