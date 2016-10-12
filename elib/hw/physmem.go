package hw

import (
	"github.com/platinasystems/go/elib"

	"unsafe"
)

var heap = &elib.MemHeap{}

func DmaInit(b []byte)                                            { heap.InitData(b) }
func DmaAlloc(n uint) (b []byte, id elib.Index, offset, cap uint) { return DmaAllocAligned(n, 0) }
func DmaFree(id elib.Index)                                       { heap.Put(id) }
func DmaGetData(id elib.Index) (b []byte)                         { return heap.GetId(id) }
func DmaGetPointer(o uint) unsafe.Pointer                         { return heap.Data(o) }
func DmaIsValidOffset(o uint) bool                                { return heap.OffsetValid(o) }
func DmaHeapUsage() string                                        { return heap.String() }
