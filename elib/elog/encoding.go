// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cpu"

	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"math"
)

func Uvarint(b []byte) (c []byte, i int) {
	x, n := binary.Uvarint(b)
	i = int(x)
	c = b[n:]
	return
}

func PutUvarint(b []byte, i int) (c []byte) {
	n := binary.PutUvarint(b, uint64(i))
	c = b[n:]
	return
}

func (v *View) Save(w io.Writer) (err error) {
	enc := gob.NewEncoder(w)
	err = enc.Encode(v)
	return
}

func (v *View) Restore(r io.Reader) (err error) {
	dec := gob.NewDecoder(r)
	err = dec.Decode(v)
	return
}

func (e *Event) encodeData(c *Context, b []byte) int { return e.getType().Encode(c, e, b) }
func (e *Event) decodeData(c *Context, b []byte) int { return e.getType().Decode(c, e, b) }

func (e *Event) encode(c *Context, b0 elib.ByteVec, eType uint16, t0 cpu.Time, i0 int) (b elib.ByteVec, t cpu.Time, i int) {
	b, i = b0, i0
	b.Validate(uint(i + 1<<log2EventBytes))
	// Encode time differences for shorter encodings.
	t = e.timestamp
	i += binary.PutUvarint(b[i:], uint64(t-t0))
	i += binary.PutUvarint(b[i:], uint64(e.callerIndex))
	i += binary.PutUvarint(b[i:], uint64(eType))
	i += binary.PutUvarint(b[i:], uint64(e.track))
	i += e.encodeData(c, b[i:])
	return
}

var (
	errUnderflow      = errors.New("decode buffer underflow")
	errStringOverflow = errors.New("decode string overflow")
)

func (e *Event) decode(c *Context, b elib.ByteVec, typeMap elib.Uint16Vec, t0 cpu.Time, i0 int) (t cpu.Time, i int, err error) {
	i, t = i0, t0
	var (
		x uint64
		n int
	)

	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	t += cpu.Time(x)
	e.timestamp = t
	i += n

	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	e.callerIndex = uint32(x)
	i += n

	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	if int(x) >= len(typeMap) {
		err = fmt.Errorf("type index out of range %d >= %d", x, len(typeMap))
		return
	}
	e.typeIndex = typeMap[x]
	i += n

	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	e.track = uint16(x)
	i += n

	i += e.decodeData(c, b[i:])
	return

short:
	err = errUnderflow
	return
}

func encodeString(s string, b0 elib.ByteVec, i0 int) (b elib.ByteVec, i int) {
	l := len(s)
	b.Validate(uint(i + binary.MaxVarintLen64 + l))
	i += binary.PutUvarint(b[i:], uint64(l))
	copy(b[i:], s)
	i += l
	return
}

func decodeString(b elib.ByteVec, i0, maxLen int) (s string, i int, err error) {
	i = i0
	var (
		x    uint64
		l, n int
	)
	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	if maxLen != 0 && x > uint64(maxLen) {
		err = errStringOverflow
		return
	}
	i += n
	l = int(x)
	if len(b[i:]) < l {
		goto short
	}
	s = string(b[i : i+l])
	return

short:
	err = errUnderflow
	return
}

func (c *CallerInfo) encode(b0 elib.ByteVec, i0 int) (b elib.ByteVec, i int) {
	b.Validate(uint(i + 3*binary.MaxVarintLen64))
	i += binary.PutUvarint(b[i:], uint64(c.PC))
	i += binary.PutUvarint(b[i:], uint64(c.Entry))
	i += binary.PutUvarint(b[i:], uint64(c.Line))
	b, i = encodeString(c.Name, b, i)
	b, i = encodeString(c.File, b, i)
	return
}

func (c *CallerInfo) decode(b elib.ByteVec, i0 int) (i int, err error) {
	var (
		x uint64
		n int
	)

	i = i0
	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	i += n
	c.PC = uintptr(x)

	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	i += n
	c.Entry = uintptr(x)

	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	i += n
	c.Line = int(x)

	const maxLen = 4 << 10
	c.Name, i, err = decodeString(b, i, maxLen)
	if err != nil {
		return
	}
	c.File, i, err = decodeString(b, i, maxLen)
	return

short:
	err = errUnderflow
	return
}

func (view *View) MarshalBinary() ([]byte, error) {
	var b elib.ByteVec

	i := 0
	bo := binary.BigEndian

	b.Validate(uint(i + 8))
	bo.PutUint64(b[i:], math.Float64bits(view.timeUnitNsecs()))
	i += 8

	b.Validate(uint(i + binary.MaxVarintLen64))
	i += binary.PutUvarint(b[i:], uint64(view.cpuStartTime))

	d, err := view.StartTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	b.Validate(uint(i + len(d) + binary.MaxVarintLen64))
	i += binary.PutUvarint(b[i:], uint64(len(d)))
	i += copy(b[i:], d)

	b.Validate(uint(i + binary.MaxVarintLen64))
	i += binary.PutUvarint(b[i:], uint64(len(view.Events)))

	// Map global event types to log local ones.
	var localTypes elib.Uint16Vec
	var globalTypes elib.Uint32Vec

	typesUsed := elib.Bitmap(0)
	for ei := range view.Events {
		e := &view.Events[ei]
		ti := uint(e.typeIndex)
		if !typesUsed.Get(ti) {
			typesUsed = typesUsed.Orx(ti)
			globalTypes.Validate(ti)
			globalTypes[ti] = uint32(len(localTypes))
			localTypes = append(localTypes, e.typeIndex)
		}
	}

	// Encode number of unique types followed by type names.
	b.Validate(uint(i + binary.MaxVarintLen64))
	i += binary.PutUvarint(b[i:], uint64(len(localTypes)))
	for x := range localTypes {
		t := getTypeByIndex(int(localTypes[x]))
		b.Validate(uint(i + binary.MaxVarintLen64 + len(t.Name)))
		i += binary.PutUvarint(b[i:], uint64(len(t.Name)))
		i += copy(b[i:], t.Name)
	}

	// Callers
	b.Validate(uint(i + binary.MaxVarintLen64))
	i += binary.PutUvarint(b[i:], uint64(len(view.callers)))
	for _, r := range view.callers {
		c := view.getCallerInfo(r.callerIndex)
		b, i = c.encode(b, i)
	}

	// String table.
	b, i = encodeString(string(view.stringTable.t), b, i)

	t := view.cpuStartTime
	for ei := range view.Events {
		e := &view.Events[ei]
		b, t, i = e.encode(view.GetContext(), b, uint16(globalTypes[e.typeIndex]), t, i)
	}

	return b[:i], nil
}

func (view *View) UnmarshalBinary(b []byte) (err error) {
	i := 0
	bo := binary.BigEndian

	view.timeUnitNsec = math.Float64frombits(bo.Uint64(b[i:]))
	i += 8

	if x, n := binary.Uvarint(b[i:]); n > 0 {
		view.cpuStartTime = cpu.Time(x)
		i += n
	} else {
		return errUnderflow
	}

	if x, n := binary.Uvarint(b[i:]); n > 0 {
		i += n
		timeLen := int(x)
		if i+timeLen > len(b) {
			return errUnderflow
		}
		err = view.StartTime.UnmarshalBinary(b[i : i+timeLen])
		if err != nil {
			return err
		}
		i += timeLen
	} else {
		return errUnderflow
	}

	if x, n := binary.Uvarint(b[i:]); n > 0 {
		l := uint(x)
		if len(view.Events) > 0 {
			view.Events = view.Events[:0]
		}
		view.Events.Resize(l)
		i += n
	} else {
		return errUnderflow
	}

	var typeMap elib.Uint16Vec

	if x, n := binary.Uvarint(b[i:]); n > 0 {
		typeMap.Resize(uint(x))
		i += n
	} else {
		return errUnderflow
	}

	for li := range typeMap {
		if x, n := binary.Uvarint(b[i:]); n > 0 {
			i += n
			nameLen := int(x)
			if i+nameLen > len(b) {
				return errUnderflow
			}
			name := string(b[i : i+nameLen])
			i += nameLen
			if tp, ok := getTypeByName(name); !ok {
				return fmt.Errorf("unknown type named `%s'", name)
			} else {
				typeMap[li] = uint16(tp.index)
			}
		} else {
			return errUnderflow
		}
	}

	// Callers
	if nCallers, n := binary.Uvarint(b[i:]); n > 0 {
		i += n
		for j := 0; j < int(nCallers); j++ {
			var c CallerInfo
			if i, err = c.decode(b, i); err != nil {
				return
			}
			view.addCallerInfo(c)
		}
	} else {
		return errUnderflow
	}

	// String table.
	{
		var s string
		if s, i, err = decodeString(b, i, 0); err != nil {
			return
		}
		view.stringTable.init(s)
	}

	t := view.cpuStartTime
	for ei := 0; ei < len(view.Events); ei++ {
		e := &view.Events[ei]
		t, i, err = e.decode(view.GetContext(), b, typeMap, t, i)
		if err != nil {
			return
		}
	}

	b = b[:i]

	return
}

func (t *EventType) MarshalBinary() ([]byte, error) {
	return []byte(t.Name), nil
}

func (t *EventType) UnmarshalBinary(data []byte) (err error) {
	n := string(data)
	if rt, ok := getTypeByName(n); ok {
		*t = *rt
	} else {
		err = errors.New("unknown type: " + n)
	}
	return
}

func EncodeUint32(b []byte, x uint32) int { return binary.PutUvarint(b, uint64(x)) }
func DecodeUint32(b []byte, i int) (uint32, int) {
	x, n := binary.Uvarint(b[i:])
	return uint32(x), i + n
}
func EncodeUint64(b []byte, x uint64) int { return binary.PutUvarint(b, uint64(x)) }
func DecodeUint64(b []byte, i int) (uint64, int) {
	x, n := binary.Uvarint(b[i:])
	return uint64(x), i + n
}
