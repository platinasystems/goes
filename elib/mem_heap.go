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

func RawMmap(addr, length, prot, flags, fd, offset uintptr) (a uintptr, b []byte, err error) {
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

// Init initializes heap with n bytes of mmap'ed anonymous memory.
func (h *MemHeap) init(b []byte, n uint) {
	if len(b) == 0 {
		var err error
		_, b, err = RawMmap(0, uintptr(n), syscall.PROT_READ|syscall.PROT_WRITE,
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
