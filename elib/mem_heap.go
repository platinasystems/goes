// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elib

import (
	"github.com/platinasystems/go/elib/cpu"

	"fmt"
	"reflect"
	"sync"
	"syscall"
	"unsafe"
)

func (x Word) RoundCacheLine() Word { return x.RoundPow2(cpu.CacheLineBytes) }
func RoundCacheLine(x Word) Word    { return x.RoundCacheLine() }

// Allocation heap of cache lines.
type MemHeap struct {
	// Protects heap get/put.
	mu sync.Mutex

	heap Heap

	once sync.Once

	// Virtual address lines returned via mmap of anonymous memory.
	data []byte
}

func munmap(a, size uintptr) (err error) {
	const flags = syscall.MAP_SHARED | syscall.MAP_ANONYMOUS | syscall.MAP_FIXED
	_, _, e := syscall.RawSyscall6(syscall.SYS_MMAP, a, size, syscall.PROT_NONE, flags, 0, 0)
	if e != 0 {
		err = fmt.Errorf("mmap PROT_NONE: %s", e)
	}
	return
}

func MmapSliceAligned(log2_size, log2_align uint, flags, prot uintptr) (a uintptr, b []byte, err error) {
	const log2_page_size = 12
	if log2_align < log2_page_size {
		log2_align = log2_page_size
	}
	size := uintptr(1) << log2_size
	align := uintptr(1) << log2_align
	a, _, e := syscall.RawSyscall6(syscall.SYS_MMAP, 0, size+align, prot, flags, 0, 0)
	if e != 0 {
		err = fmt.Errorf("mmap: %s", e)
		return
	}
	if align > log2_page_size {
		a0 := a
		a1 := (a0 + align - 1) &^ (align - 1)
		a2 := a1 + size
		a3 := a0 + size + align
		if a1 > a0 {
			if err = munmap(a0, a1-a0); err != nil {
				return
			}
		}
		if a3 > a2 {
			if err = munmap(a2, a3-a2); err != nil {
				return
			}
		}
		a = a1
	}
	slice := reflect.SliceHeader{Data: a, Len: int(size), Cap: int(size)}
	b = *(*[]byte)(unsafe.Pointer(&slice))
	return
}

func MmapSlice(addr, length, prot, flags, fd, offset uintptr) (a uintptr, b []byte, err error) {
	r, _, e := syscall.RawSyscall6(syscall.SYS_MMAP, addr, length, prot, flags, fd, offset)
	if e != 0 {
		err = fmt.Errorf("mmap: %s", e)
		return
	}
	slice := reflect.SliceHeader{Data: r, Len: int(length), Cap: int(length)}
	a = r
	b = *(*[]byte)(unsafe.Pointer(&slice))
	return
}

func Munmap(b []byte) (err error) {
	_, _, e := syscall.RawSyscall(syscall.SYS_MUNMAP, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), 0)
	if e != 0 {
		err = fmt.Errorf("munmap: %s", e)
	}
	return
}

// Init initializes heap with n bytes of mmap'ed anonymous memory.
func (h *MemHeap) init(b []byte, n uint) {
	if len(b) == 0 {
		var err error
		_, b, err = MmapSlice(0, uintptr(n), syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_PRIVATE|syscall.MAP_ANON|syscall.MAP_NORESERVE, 0, 0)
		if err != nil {
			err = fmt.Errorf("mmap: %s", err)
			panic(err)
		}
	}
	n = uint(len(b)) &^ (cpu.CacheLineBytes - 1)
	h.data = b[:n]
	h.heap.SetMaxLen(n >> cpu.Log2CacheLineBytes)
}

func (h *MemHeap) Init(n uint) (err error) {
	h.once.Do(func() { h.init(h.data, n) })
	return
}

func (h *MemHeap) InitData(b []byte) { h.init(b, 0) }

func (h *MemHeap) GetAligned(n, log2Align uint) (b []byte, id Index, offset, cap uint) {
	// Allocate memory in case caller has not called Init to select a size.
	if err := h.Init(64 << 20); err != nil {
		panic(err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if log2Align < cpu.Log2CacheLineBytes {
		log2Align = cpu.Log2CacheLineBytes
	}
	log2Align -= cpu.Log2CacheLineBytes

	cap = uint(Word(n).RoundCacheLine())
	id, i := h.heap.GetAligned(cap>>cpu.Log2CacheLineBytes, log2Align)
	offset = uint(i) << cpu.Log2CacheLineBytes
	b = h.data[offset : offset+cap]
	return
}

func (h *MemHeap) Get(n uint) (b []byte, id Index, offset, cap uint) { return h.GetAligned(n, 0) }

func (h *MemHeap) Put(id Index) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.heap.Put(id)
}

func (h *MemHeap) GetId(id Index) (b []byte) {
	offset, len := h.heap.GetID(id)
	return h.data[offset : offset+len]
}

func (h *MemHeap) Offset(b []byte) uint {
	return uint(uintptr(unsafe.Pointer(&b[0])) - uintptr(unsafe.Pointer(&h.data[0])))
}

func (h *MemHeap) Data(o uint) unsafe.Pointer { return unsafe.Pointer(&h.data[o]) }
func (h *MemHeap) OffsetValid(o uint) bool    { return o < uint(len(h.data)) }

func (h *MemHeap) String() string {
	max := h.heap.GetMaxLen()
	if max == 0 {
		return "empty"
	}
	u := h.heap.GetUsage()
	return fmt.Sprintf("used %s, free %s, capacity %s",
		MemorySize(u.Used<<cpu.Log2CacheLineBytes),
		MemorySize(u.Free<<cpu.Log2CacheLineBytes),
		MemorySize(max<<cpu.Log2CacheLineBytes))
}
