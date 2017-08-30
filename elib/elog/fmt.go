// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"github.com/platinasystems/go/elib"

	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strings"
)

// Generic events using fmt.Printf formatting.
type fmtKind uint8

const (
	fmtEnd = iota
	fmtBoolTrue
	fmtBoolFalse
	fmtUint
	fmtInt
	fmtFloat
	fmtStringRef
	fmtString
)

type fmtEvent struct {
	b [EventDataBytes]byte
}

func (e *fmtEvent) Elog(l *Log) {
	format, args := e.decode(l)
	l.Logf(format, args...)
}

func encodeInt(b *elib.ByteVec, i0 uint, v int64) (i uint) {
	i = i0
	b.Validate(i + 1 + binary.MaxVarintLen64)
	(*b)[i] = fmtInt
	i += 1 + uint(binary.PutVarint((*b)[i+1:], v))
	return
}

func encodeUintb(b []byte, i0 uint, v uint64, kind int) (i uint) {
	i = i0
	b[i] = byte(kind)
	i += 1 + uint(binary.PutUvarint(b[i+1:], v))
	return
}

func encodeUint(b *elib.ByteVec, i0 uint, v uint64, kind int) (i uint) {
	i = i0
	b.Validate(i + 1 + binary.MaxVarintLen64)
	i = encodeUintb(*b, i, v, kind)
	return
}

func encodeBool(b *elib.ByteVec, i0 uint, v bool) (i uint) {
	i = i0
	b.Validate(i + 1)
	(*b)[i] = fmtBoolFalse
	if v {
		(*b)[i] = fmtBoolTrue
	}
	i++
	return
}

func encodeStr(b *elib.ByteVec, i0 uint, v string) (i uint) {
	l := uint(len(v))
	b.Validate(i + 2 + l)
	(*b)[i] = fmtString
	(*b)[i+1] = byte(l)
	copy((*b)[i+2:], v)
	i += 2 + l
	return
}

func (s *shared) encodeArg(b *elib.ByteVec, i0 uint, a interface{}) (i uint) {
	i = i0
	switch v := a.(type) {
	case bool:
		i = encodeBool(b, i, v)
	case string:
		i = encodeStr(b, i, v)
	case StringRef:
		i = encodeUint(b, i, uint64(v), fmtStringRef)
	case int8:
		i = encodeInt(b, i, int64(v))
	case int16:
		i = encodeInt(b, i, int64(v))
	case int32:
		i = encodeInt(b, i, int64(v))
	case int64:
		i = encodeInt(b, i, int64(v))
	case int:
		i = encodeInt(b, i, int64(v))
	case uint8:
		i = encodeUint(b, i, uint64(v), fmtUint)
	case uint16:
		i = encodeUint(b, i, uint64(v), fmtUint)
	case uint32:
		i = encodeUint(b, i, uint64(v), fmtUint)
	case uint64:
		i = encodeUint(b, i, uint64(v), fmtUint)
	case uint:
		i = encodeUint(b, i, uint64(v), fmtUint)
	case float64:
		i = encodeUint(b, i, uint64(math.Float64bits(v)), fmtFloat)
	case float32:
		i = encodeUint(b, i, uint64(float64(math.Float32bits(v))), fmtFloat)
	default:
		// Convert String() to index into string table and save for re-use.
		if r, ok := a.(fmt.Stringer); ok {
			i = encodeUint(b, i, uint64(s.SetString(r.String())), fmtStringRef)
		} else {
			val := reflect.ValueOf(a)
			switch val.Kind() {
			case reflect.Bool:
				i = encodeBool(b, i, val.Bool())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				i = encodeInt(b, i, val.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				i = encodeUint(b, i, val.Uint(), fmtUint)
			case reflect.String:
				i = encodeStr(b, i, val.String())
			default:
				panic(fmt.Errorf("elog fmtEvent encode value with unknown type: %v", val.Type()))
			}
		}
	}
	return
}

func fmtEncode(s *shared, b *elib.ByteVec, i0 uint, doArgs bool, e *fmtEvent, r StringRef, format string, args []interface{}) (f StringRef, i uint) {
	i = i0
	b.Validate(i + binary.MaxVarintLen64)

	f = r
	if f == StringRefNil {
		f = s.SetString(strings.TrimSpace(format))
	}
	i += uint(binary.PutUvarint((*b)[i:], uint64(f)))

	if doArgs {
		for _, a := range args {
			i = s.encodeArg(b, i, a)
		}
		if e == nil || i+1 < uint(len(e.b)) {
			b.Validate(i + 1)
			(*b)[i] = fmtEnd
			i++
		}
	}
	if e != nil {
		if uint(copy(e.b[:], (*b)[i0:i])) < i-i0 {
			panic("overflow")
		}
	}
	*b = (*b)[:i]
	return
}

func (e *fmtEvent) encode(s *shared, doArgs bool, r StringRef, format string, args []interface{}) (f StringRef, i uint) {
	s.fmtMu.Lock()
	defer s.fmtMu.Unlock()
	if s.fmtBuffer != nil {
		s.fmtBuffer = s.fmtBuffer[:0]
	}
	return fmtEncode(s, &s.fmtBuffer, 0, doArgs, e, r, format, args)
}

func (s *shared) decodeArg(b []byte, i0 int) (a interface{}, kind byte, i int) {
	i = i0
	if i >= len(b) {
		kind = fmtEnd
		return
	}
	kind = b[i]
	i++
	switch kind {
	case fmtEnd:
	case fmtBoolTrue, fmtBoolFalse:
		a = kind == fmtBoolTrue
	case fmtInt:
		x, n := binary.Varint(b[i:])
		i += n
		a = x
	case fmtUint:
		x, n := binary.Uvarint(b[i:])
		i += n
		a = x
	case fmtFloat:
		x, n := binary.Uvarint(b[i:])
		i += n
		a = math.Float64frombits(x)
	case fmtStringRef:
		x, n := binary.Uvarint(b[i:])
		i += n
		a = s.GetString(StringRef(x))
	case fmtString:
		l := int(b[i])
		a = string(b[i+1 : i+1+l])
		i += 1 + l
	default:
		panic(fmt.Errorf("elog fmtEvent decode unknown kind: 0x%x", kind))
	}
	return
}

func (e *fmtEvent) decode(l *Log) (format string, args []interface{}) {
	b := e.b[:]
	i := 0

	{
		x, n := binary.Uvarint(b[i:])
		i += n
		format = l.s.GetString(StringRef(x))
	}

	for {
		var (
			a    interface{}
			kind byte
		)
		if a, kind, i = l.s.decodeArg(b, i); kind == fmtEnd {
			break
		} else {
			args = append(args, a)
		}
	}
	return
}

func (b *Buffer) fmt(c Caller, t uint, format string, args []interface{}) {
	if !Enabled() {
		return
	}
	if !b.Enabled() {
		return
	}
	if r, disabled := b.getCaller(nil, c); !disabled {
		f := &r.fe
		r.fmtIndex, _ = f.encode(&b.shared, true, r.fmtIndex, format, args)
		b.add1(f, c, t, r)
	}
}

func (b *Buffer) F(format string, args ...interface{}) {
	c := b.GetCaller(PointerToFirstArg(&format))
	b.fmt(c, 0, format, args)
}
func (b *Buffer) Fc(format string, c Caller, args ...interface{}) {
	b.fmt(c, 0, format, args)
}
func F(format string, args ...interface{}) {
	b := DefaultBuffer
	c := b.GetCaller(PointerToFirstArg(&format))
	b.fmt(c, 0, format, args)
}
func Fc(format string, c Caller, args ...interface{}) {
	DefaultBuffer.fmt(c, 0, format, args)
}

func (b *Buffer) S(s string) {
	c := b.GetCaller(PointerToFirstArg(&b))
	b.fmt(c, 0, s, nil)
}
func (b *Buffer) Sc(s string, c Caller) {
	b.fmt(c, 0, s, nil)
}
func S(s string) {
	b := DefaultBuffer
	c := b.GetCaller(PointerToFirstArg(&s))
	b.Sc(s, c)
}
func Sc(s string, c Caller) {
	DefaultBuffer.Sc(s, c)
}

//go:generate go run genfmt.go

func (b *Buffer) F1b(format string, v bool) {
	c := b.GetCaller(PointerToFirstArg(&b))
	b.Fc1b(format, c, v)
}
func (b *Buffer) Fc1b(format string, c Caller, v bool) {
	if !Enabled() {
		return
	}
	if r, disabled := b.getCaller(nil, c); !disabled {
		f := &r.fe
		var i uint
		r.fmtIndex, i = f.encode(&b.shared, false, r.fmtIndex, format, nil)
		bv := byte(fmtBoolFalse)
		if v {
			bv = fmtBoolTrue
		}
		f.b[i] = bv
		f.b[i+1] = fmtEnd
		b.add1(f, c, 0, r)
	}
}
func F1b(f string, v bool) {
	b := DefaultBuffer
	c := b.GetCaller(PointerToFirstArg(&f))
	b.Fc1b(f, c, v)
}
func Fc1b(f string, c Caller, v bool) { DefaultBuffer.Fc1b(f, c, v) }

func (b *Buffer) F1u(format string, v uint64) {
	c := b.GetCaller(PointerToFirstArg(&format))
	b.Fc1u(format, c, v)
}
func (b *Buffer) Fc1u(format string, c Caller, v uint64) {
	if !Enabled() {
		return
	}
	if r, disabled := b.getCaller(nil, c); !disabled {
		f := &r.fe
		var i uint
		r.fmtIndex, i = f.encode(&b.shared, false, r.fmtIndex, format, nil)
		i = encodeUintb(f.b[:], i, v, fmtUint)
		f.b[i+1] = fmtEnd
		b.add1(f, c, 0, r)
	}
}
func F1u(f string, v uint64) {
	b := DefaultBuffer
	c := b.GetCaller(PointerToFirstArg(&f))
	b.Fc1u(f, c, v)
}
func Fc1u(f string, c Caller, v uint64) { DefaultBuffer.Fc1u(f, c, v) }

func (b *Buffer) F2u(format string, v0, v1 uint64) {
	c := b.GetCaller(PointerToFirstArg(&format))
	b.Fc2u(format, c, v0, v1)
}
func (b *Buffer) Fc2u(format string, c Caller, v0, v1 uint64) {
	if !Enabled() {
		return
	}
	if r, disabled := b.getCaller(nil, c); !disabled {
		f := &r.fe
		var i uint
		r.fmtIndex, i = f.encode(&b.shared, false, r.fmtIndex, format, nil)
		i = encodeUintb(f.b[:], i, v0, fmtUint)
		i = encodeUintb(f.b[:], i, v1, fmtUint)
		f.b[i+1] = fmtEnd
		b.add1(f, c, 0, r)
	}
}
func F2u(f string, v0, v1 uint64) {
	b := DefaultBuffer
	c := b.GetCaller(PointerToFirstArg(&f))
	b.Fc2u(f, c, v0, v1)
}
func Fc2u(f string, c Caller, v0, v1 uint64) { DefaultBuffer.Fc2u(f, c, v0, v1) }
