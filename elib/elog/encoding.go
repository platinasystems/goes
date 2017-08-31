// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"github.com/platinasystems/go/elib"

	"encoding/binary"
	"encoding/gob"
	"errors"
	"io"
	"math"
	"os"
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

func SaveView(file string) (err error) {
	var f *os.File
	if f, err = os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0666); err != nil {
		return
	}
	defer f.Close()
	v := NewView()
	v.SetName(file)
	err = v.Save(f)
	return
}

func (v *View) Load(file string) (err error) {
	var f *os.File
	if f, err = os.OpenFile(file, os.O_RDONLY, 0); err != nil {
		return
	}
	defer f.Close()
	v.Restore(f)
	return
}

func (e *viewEvent) encode(b0 elib.ByteVec, t0 uint64, i0 int) (b elib.ByteVec, t uint64, i int) {
	b, i = b0, i0
	b.Validate(uint(i + 5*binary.MaxVarintLen64))
	// Encode time differences for shorter encodings.
	t = e.timestamp
	i += binary.PutUvarint(b[i:], uint64(t-t0))
	i += binary.PutUvarint(b[i:], uint64(e.callerIndex))
	i += binary.PutUvarint(b[i:], uint64(e.trackIndex))
	i += binary.PutUvarint(b[i:], uint64(e.lo))
	i += binary.PutUvarint(b[i:], uint64(e.hi))
	return
}

var (
	errUnderflow      = errors.New("decode buffer underflow")
	errStringOverflow = errors.New("decode string overflow")
)

func (e *viewEvent) decode(b elib.ByteVec, t0 uint64, i0 int) (t uint64, i int, err error) {
	i, t = i0, t0
	var (
		x uint64
		n int
	)

	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	t += uint64(x)
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
	e.trackIndex = uint32(x)
	i += n

	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	e.lo = uint32(x)
	i += n

	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	e.hi = uint32(x)
	i += n
	return

short:
	err = errUnderflow
	return
}

func encodeString(b0 elib.ByteVec, i0 int, s string) (b elib.ByteVec, i int) {
	b, i = b0, i0
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
	i += l
	return

short:
	err = errUnderflow
	return
}

func (c *CallerInfo) encode(b0 elib.ByteVec, i0 int) (b elib.ByteVec, i int) {
	b, i = b0, i0
	b.Validate(uint(i + 3*binary.MaxVarintLen64))
	i += binary.PutUvarint(b[i:], uint64(c.PC))
	i += binary.PutUvarint(b[i:], uint64(c.Entry))
	i += binary.PutUvarint(b[i:], uint64(c.Line))
	b, i = encodeString(b, i, c.Name)
	b, i = encodeString(b, i, c.File)
	return
}

func (c *CallerInfo) decode(b elib.ByteVec, i0 int) (i int, err error) {
	i = i0

	var (
		x uint64
		n int
	)
	const maxLen = 4 << 10
	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	i += n
	c.PC = x

	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	i += n
	c.Entry = x

	if x, n = binary.Uvarint(b[i:]); n <= 0 {
		goto short
	}
	i += n
	c.Line = int(x)

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

func (v *View) MarshalBinary() ([]byte, error) {
	var b elib.ByteVec

	i := 0
	bo := binary.BigEndian

	// Name
	{
		l := len(v.name)
		b.Validate(uint(i + l + binary.MaxVarintLen64))
		i += binary.PutUvarint(b[i:], uint64(l))
		copy(b[i:], v.name)
		i += l
	}

	// Header
	b.Validate(uint(i + 8))
	bo.PutUint64(b[i:], math.Float64bits(v.cpuTimeUnitNsec))
	i += 8

	b.Validate(uint(i + binary.MaxVarintLen64))
	i += binary.PutUvarint(b[i:], uint64(v.cpuStartTime))

	d, err := v.StartTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	b.Validate(uint(i + len(d) + binary.MaxVarintLen64))
	i += binary.PutUvarint(b[i:], uint64(len(d)))
	i += copy(b[i:], d)

	// Callers
	v.convertBufferEvents()
	b.Validate(uint(i + binary.MaxVarintLen64))
	i += binary.PutUvarint(b[i:], uint64(len(v.callers)))
	for _, r := range v.callers {
		_, c := v.getCallerInfo(r.callerIndex)
		b, i = c.encode(b, i)
	}

	// String table.
	b, i = encodeString(b, i, string(v.stringTable.strings))

	// Events.
	b.Validate(uint(i + binary.MaxVarintLen64))
	i += binary.PutUvarint(b[i:], uint64(len(v.allViewEvents)))
	t := v.cpuStartTime
	for ei := range v.allViewEvents {
		e := &v.allViewEvents[ei]
		b, t, i = e.encode(b, t, i)
	}

	// Event formats and arguments.
	b, i = encodeString(b, i, string(v.viewEvents.b))

	return b[:i], nil
}

func (v *View) UnmarshalBinary(b []byte) (err error) {
	i := 0
	bo := binary.BigEndian

	// Name
	if x, n := binary.Uvarint(b[i:]); n > 0 {
		l := int(x)
		i += n
		v.name = string(b[i : i+l])
		i += l
	} else {
		return errUnderflow
	}

	v.cpuTimeUnitNsec = math.Float64frombits(bo.Uint64(b[i:]))
	i += 8

	if x, n := binary.Uvarint(b[i:]); n > 0 {
		v.cpuStartTime = uint64(x)
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
		err = v.StartTime.UnmarshalBinary(b[i : i+timeLen])
		if err != nil {
			return err
		}
		i += timeLen
	} else {
		return errUnderflow
	}

	// Callers
	if nCallers, n := binary.Uvarint(b[i:]); n > 0 {
		i += n
		for j := 0; j < int(nCallers); j++ {
			var c CallerInfo
			if i, err = c.decode(b, i); err != nil {
				return
			}
			v.addCallerInfo(c)
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
		v.stringTable.init(s)
	}

	// Events.
	if x, n := binary.Uvarint(b[i:]); n > 0 {
		l := uint(x)
		if len(v.allViewEvents) > 0 {
			v.allViewEvents = v.allViewEvents[:0]
		}
		v.allViewEvents = make([]viewEvent, l)
		i += n
	} else {
		return errUnderflow
	}
	t := v.cpuStartTime
	for ei := 0; ei < len(v.allViewEvents); ei++ {
		e := &v.allViewEvents[ei]
		t, i, err = e.decode(b, t, i)
		if err != nil {
			return
		}
	}

	// Event formats and arguments.
	{
		var s string
		if s, i, err = decodeString(b, i, 0); err != nil {
			return
		}
		v.viewEvents.b = []byte(s)
	}

	v.currentViewEvents = v.allViewEvents
	v.getViewTimes()

	b = b[:i]
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
