// +build uio_pci_dma

package hw

import (
	"github.com/platinasystems/go/elib"
)

type pageTable struct {
	Data uintptr

	Pages []uintptr

	Log2BytesPerPage uint

	free []elib.Index
}

var PageTable pageTable

func DmaPhysAddress(a uintptr) uintptr {
	t := &PageTable
	l := t.Log2BytesPerPage
	o := a - t.Data
	return t.Pages[o>>l] + o&(1<<l-1)
}

func (t *pageTable) samePage(o, n uint) bool {
	l := t.Log2BytesPerPage
	return o>>l == (o+n-1)>>l
}

func DmaAllocAligned(n, log2Align uint) (b []byte, id elib.Index, offset, cap uint) {
	t := &PageTable
	t.free = t.free[:0]
	for {
		b, id, offset, cap = heap.GetAligned(n, log2Align)
		// Reject allocations that span page boundaries.
		if t.samePage(offset, n) {
			break
		}
		t.free = append(t.free, id)
	}
	for _, x := range t.free {
		heap.Put(x)
	}
	return
}
