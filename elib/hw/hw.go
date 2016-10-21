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
	RegsBasePointer = basePointer()
	RegsBaseAddress = uintptr(RegsBasePointer)
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
func LoadUint32(addr *uint32) (data uint32)
func StoreUint32(addr *uint32, data uint32)
func LoadUint64(addr *uint64) (data uint64)
func StoreUint64(addr *uint64, data uint64)

func MemoryBarrier()

// Generic 8/16/32 bit registers
type Reg8 uint8
type Reg16 uint16
type Reg32 uint32

// Byte offsets
func (r *Reg8) Offset() uint  { return uint(uintptr(unsafe.Pointer(r)) - RegsBaseAddress) }
func (r *Reg16) Offset() uint { return uint(uintptr(unsafe.Pointer(r)) - RegsBaseAddress) }
func (r *Reg32) Offset() uint { return uint(uintptr(unsafe.Pointer(r)) - RegsBaseAddress) }

func (r *Reg32) Get() uint32  { return LoadUint32((*uint32)(r)) }
func (r *Reg32) Set(x uint32) { StoreUint32((*uint32)(r), x) }
