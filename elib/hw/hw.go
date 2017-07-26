// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Memory mapped register read/write
package hw

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Must point to readable memory since compiler may perform
// read probes (nil checks) as part of memory addressing.
var (
	BasePointer = basePointer()
	BaseAddress = uintptr(BasePointer)
)

func basePointer() unsafe.Pointer {
	// ok for all 32 bit devices.
	x, err := syscall.Mmap(0, 0, 1<<32, syscall.PROT_READ, syscall.MAP_PRIVATE|syscall.MAP_ANON|syscall.MAP_NORESERVE)
	if err != nil {
		panic(err)
	}
	return unsafe.Pointer(&x[0])
}

func CheckRegAddr(name string, got, want uint) {
	if got != want {
		panic(fmt.Errorf("%s got 0x%x != want 0x%x", name, got, want))
	}
}

// Memory-mapped read/write
func LoadUint32(addr uintptr) (data uint32)
func StoreUint32(addr uintptr, data uint32)
func LoadUint64(addr uintptr) (data uint64)
func StoreUint64(addr uintptr, data uint64)

func MemoryBarrier()

// Generic 8/16/32 bit registers
type U8 uint8
type U16 uint16
type U32 uint32

// Byte offsets
func (r *U8) Offset() uintptr  { return uintptr(unsafe.Pointer(r)) - BaseAddress }
func (r *U16) Offset() uintptr { return uintptr(unsafe.Pointer(r)) - BaseAddress }
func (r *U32) Offset() uintptr { return uintptr(unsafe.Pointer(r)) - BaseAddress }

func (r *U32) Get(base uintptr) uint32    { return LoadUint32(base + r.Offset()) }
func (r *U32) Set(base uintptr, x uint32) { StoreUint32(base+r.Offset(), x) }
