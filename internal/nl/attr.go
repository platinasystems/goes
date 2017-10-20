// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nl

import (
	"io"
	"syscall"
	"unsafe"

	"github.com/platinasystems/go/internal/safe"
)

const SizeofRtAttr = syscall.SizeofRtAttr

func ForEachAttr(b []byte, do func(uint16, []byte)) {
	for i := 0; i <= len(b)-SizeofRtAttr; {
		h := (*syscall.RtAttr)(unsafe.Pointer(&b[i]))
		l := int(h.Len)
		n := i + l
		if l < SizeofRtAttr || n > len(b) {
			break
		}
		do(h.Type, b[i+SizeofRtAttr:n])
		i = NLATTR.Align(n)
	}
}

// Parse attribute list to index by type.
// Use IndexByType(a, Empty) to de-reference attribute data.
func IndexAttrByType(a [][]byte, b []byte) {
	if len(b) == 0 {
		for i := range a {
			a[i] = Empty
		}
	} else {
		ForEachAttr(b, func(t uint16, val []byte) {
			if t < uint16(len(a)) {
				a[t] = val
			}
		})
	}
}

func Int8(b []byte) int8 {
	if len(b) >= 1 {
		return int8(b[0])
	}
	return 0
}

func Int16(b []byte) int16 {
	if len(b) >= 2 {
		return *(*int16)(unsafe.Pointer(&b[0]))
	}
	return 0
}

func Int32(b []byte) int32 {
	if len(b) >= 4 {
		return *(*int32)(unsafe.Pointer(&b[0]))
	}
	return 0
}

func Int64(b []byte) int64 {
	if len(b) >= 8 {
		return *(*int64)(unsafe.Pointer(&b[0]))
	}
	return 0
}

func Kstring(b []byte) string {
	if len(b) >= 1 {
		return string(b[:len(b)-1])
	}
	return ""
}

func Uint8(b []byte) (v uint8) {
	if len(b) >= 1 {
		return uint8(b[0])
	}
	return 0
}

func Uint16(b []byte) (v uint16) {
	if len(b) >= 2 {
		return *(*uint16)(unsafe.Pointer(&b[0]))
	}
	return 0
}

func Uint32(b []byte) (v uint32) {
	if len(b) >= 4 {
		return *(*uint32)(unsafe.Pointer(&b[0]))
	}
	return 0
}

func Uint64(b []byte) uint64 {
	if len(b) >= 8 {
		return *(*uint64)(unsafe.Pointer(&b[0]))
	}
	return 0
}

func ReadAllAttrs(b []byte, attrs ...Attr) (int, error) {
	var n int
	for _, attr := range attrs {
		na, err := attr.Read(b[n:])
		if err != nil {
			return n, err
		}
		n += na
	}
	return n, nil
}

type Attr struct {
	Type  uint16
	Value io.Reader
}

func (attr Attr) Read(b []byte) (int, error) {
	rtattr := (*syscall.RtAttr)(unsafe.Pointer(&b[0]))
	n, err := attr.Value.Read(b[syscall.SizeofRtAttr:])
	if err != nil {
		return 0, err
	}
	*rtattr = syscall.RtAttr{
		Len:  uint16(SizeofRtAttr + n),
		Type: attr.Type,
	}
	return NLATTR.Align(syscall.SizeofRtAttr + n), nil
}

type Attrs []Attr

func (attrs Attrs) Read(b []byte) (int, error) {
	var i int

	for _, attr := range attrs {
		n, err := attr.Read(b[i:])
		if err != nil {
			return n, err
		}
		i += n
	}
	return i, nil
}

type NilAttr struct{}
type BytesAttr []byte
type Int8Attr int8
type Int16Attr int16
type Int32Attr int32
type Int64Attr int64
type KstringAttr string
type Uint8Attr uint8
type Uint16Attr uint16
type Uint32Attr uint32
type Uint64Attr uint64

// big-endian
type Be16Attr uint16
type Be32Attr uint32
type Be64Attr uint64

func (v NilAttr) Read(b []byte) (int, error) { return 0, nil }

func (v BytesAttr) Read(b []byte) (int, error) { return safe.Cp(b, v) }

func (v Int8Attr) Read(b []byte) (int, error) {
	if len(b) < 1 {
		return 0, syscall.EOVERFLOW
	}
	b[0] = byte(v)
	return 1, nil
}

func (v Int16Attr) Read(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, syscall.EOVERFLOW
	}
	*(*int16)(unsafe.Pointer(&b[0])) = int16(v)
	return 2, nil
}

func (v Int32Attr) Read(b []byte) (int, error) {
	if len(b) < 4 {
		return 0, syscall.EOVERFLOW
	}
	*(*int32)(unsafe.Pointer(&b[0])) = int32(v)
	return 4, nil
}

func (v Int64Attr) Read(b []byte) (int, error) {
	if len(b) < 8 {
		return 0, syscall.EOVERFLOW
	}
	*(*int64)(unsafe.Pointer(&b[0])) = int64(v)
	return 8, nil
}

func (v KstringAttr) Read(b []byte) (int, error) {
	if len(b) < len(v)+1 {
		return 0, syscall.EOVERFLOW
	}
	copy(b, []byte(v))
	b[len(v)] = 0
	return len(v) + 1, nil
}

func (v Uint8Attr) Read(b []byte) (int, error) {
	if len(b) < 1 {
		return 0, syscall.EOVERFLOW
	}
	b[0] = byte(v)
	return 1, nil
}

func (v Uint16Attr) Read(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, syscall.EOVERFLOW
	}
	*(*uint16)(unsafe.Pointer(&b[0])) = uint16(v)
	return 2, nil
}

func (v Uint32Attr) Read(b []byte) (int, error) {
	if len(b) < 4 {
		return 0, syscall.EOVERFLOW
	}
	*(*uint32)(unsafe.Pointer(&b[0])) = uint32(v)
	return 4, nil
}

func (v Uint64Attr) Read(b []byte) (int, error) {
	if len(b) < 8 {
		return 0, syscall.EOVERFLOW
	}
	*(*uint64)(unsafe.Pointer(&b[0])) = uint64(v)
	return 8, nil
}

func (v Be16Attr) Read(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, syscall.EOVERFLOW
	}
	b[0] = byte(v >> 8)
	b[1] = byte(v)
	return 2, nil
}

func (v Be32Attr) Read(b []byte) (int, error) {
	if len(b) < 4 {
		return 0, syscall.EOVERFLOW
	}
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
	return 4, nil
}

func (v Be64Attr) Read(b []byte) (int, error) {
	if len(b) < 8 {
		return 0, syscall.EOVERFLOW
	}
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
	return 8, nil
}
