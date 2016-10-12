// +build !uio_pci_dma

package hw

import (
	"github.com/platinasystems/go/elib"
)

func DmaAllocAligned(n, log2Align uint) (b []byte, id elib.Index, offset, cap uint) {
	return heap.GetAligned(n, log2Align)
}
func DmaPhysAddress(a uintptr) uintptr { return a }
