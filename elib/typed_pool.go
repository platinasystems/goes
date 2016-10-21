// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elib

import (
	"fmt"
	"reflect"
	"unsafe"
)

type TypedPool struct {
	// Allocation free list and free bitmap.
	pool Pool

	// Object type and data vectors.
	object_types ByteVec
	object_data  ByteVec

	// Size of largest type in bytes.
	object_size uint32

	// Unused data is poisoned; new data is zeroed when allocated.
	poison, zero    []byte
	poison_zero_buf [128]byte
}

type TypedPoolIndex uint32
type TypedPoolType uint32

func (p *TypedPool) Init(args ...interface{}) {
	for i := range args {
		if x := uint32(reflect.TypeOf(args[i]).Size()); x > p.object_size {
			p.object_size = x
		}
	}

	l := int(p.object_size)
	b := p.poison_zero_buf[:]
	if 2*l > len(b) {
		b = make([]byte, 2*l)
	}
	p.zero = b[0:l]
	p.poison = b[l : 2*l]
	dead := [...]byte{0xde, 0xad}
	for i := range p.poison {
		p.poison[i] = dead[i%len(dead)]
	}
}

func (p *TypedPool) IsInitialized() bool { return p.object_size != 0 }

func (p *TypedPool) GetIndex(typ TypedPoolType) (i TypedPoolIndex) {
	i = TypedPoolIndex(p.pool.GetIndex(uint(len(p.object_types))))
	p.object_types.Validate(uint(i))
	p.object_types[i] = byte(typ)
	s := uint(p.object_size)
	j := uint(i) * s
	p.object_data.Validate(uint(j + s - 1))
	copy(p.object_data[j:j+s], p.zero)
	return
}

func (p *TypedPool) PutIndex(t TypedPoolType, i TypedPoolIndex) (ok bool) {
	ok = p.object_types[i] == byte(t)
	if !ok {
		return
	}
	ok = p.pool.PutIndex(uint(i))
	if !ok {
		return
	}
	p.object_types[i] = 0
	s := uint(p.object_size)
	j := uint(i) * s
	copy(p.object_data[j:j+s], p.poison)
	return
}

func (p *TypedPool) GetData(t TypedPoolType, i TypedPoolIndex) unsafe.Pointer {
	if want := TypedPoolType(p.object_types[i]); want != t {
		panic(fmt.Errorf("wrong type want %d != got %d", want, t))
	}
	return unsafe.Pointer(&p.object_data[uint32(i)*p.object_size])
}

func (p *TypedPool) Data(i TypedPoolIndex) (t TypedPoolType, x unsafe.Pointer) {
	t = TypedPoolType(p.object_types[i])
	s := uint(p.object_size)
	j := uint(i) * s
	x = unsafe.Pointer(&p.object_data[j])
	return
}

func (p *TypedPool) IsFree(i uint) (ok bool) { return p.pool.IsFree(i) }
func (p *TypedPool) FreeLen() uint           { return p.pool.FreeLen() }
func (p *TypedPool) MaxLen() uint            { return p.pool.MaxLen() }
func (p *TypedPool) SetMaxLen(x uint)        { p.pool.SetMaxLen(x) }
