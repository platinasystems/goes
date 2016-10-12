package elib

import (
	"fmt"
)

// Index gives common type for indices in Heaps, Pools, Fifos, ...
type Index uint32

const MaxIndex Index = ^Index(0)

// A Heap maintains an allocator for arbitrary sized blocks of an underlying array.
// The array is not part of the Heap.
type Heap struct {
	elts []heapElt

	// Slices of free elts indices indexed by size.
	// "Size" 0 is for large sized chunks.
	free freeEltsVec

	removed []Index

	head, tail Index

	// Total number of indices allocated
	len Index

	// Largest size ever allocated
	maxSize Index

	// Max limit on heap size in elements.
	maxLen Index
}

func (heap *Heap) SetMaxLen(l uint) {
	heap.maxLen = Index(l)
}

func (heap *Heap) GetMaxLen() uint { return uint(heap.maxLen) }

type HeapUsage struct {
	Used, Free uint64
}

func (heap *Heap) GetUsage() (u HeapUsage) {
	for i := range heap.elts {
		e := &heap.elts[i]
		size := uint64(heap.eltSize(e))
		if e.isFree() {
			u.Free += size
		} else {
			u.Used += size
		}
	}
	return
}

type freeElt Index

//go:generate gentemplate -d Package=elib -id freeElt -d VecType=freeEltVec -d Type=freeElt vec.tmpl
//go:generate gentemplate -d Package=elib -id freeElts -d VecType=freeEltsVec -d Type=freeEltVec vec.tmpl

type heapElt struct {
	// Offset of this element in heap.
	offset Index

	// Index on free list for this size or ^uint32(0) if not free.
	free Index

	// Index of next and previous elements
	next, prev Index
}

func (e *heapElt) isFree() bool {
	return e.free != MaxIndex
}

func (heap *Heap) freeAfter(ei, eSize, freeSize Index) {
	// Fetch elt and new free elt.
	fi := heap.newElt()
	e, f := &heap.elts[ei], &heap.elts[fi]

	f.offset = e.offset + Index(eSize-freeSize)
	f.next = e.next
	f.prev = ei
	if f.next != MaxIndex {
		heap.elts[f.next].prev = fi
	}

	e.next = fi
	if ei == heap.tail {
		heap.tail = fi
	}
	heap.freeElt(fi, freeSize)
}

func (heap *Heap) freeBefore(ei, eSize, freeSize Index) {
	// Fetch elt and new free elt.
	fi := heap.newElt()
	e, f := &heap.elts[ei], &heap.elts[fi]

	f.offset = e.offset
	f.prev = e.prev
	f.next = ei
	if f.prev == MaxIndex {
		heap.head = fi
	} else {
		heap.elts[f.prev].next = fi
	}

	e.offset += freeSize
	e.prev = fi
	heap.freeElt(fi, freeSize)
}

func (heap *Heap) freeElt(ei, size Index) {
	if size > heap.maxSize {
		size = 0
	}
	heap.free.Validate(uint(size))
	heap.elts[ei].free = Index(len(heap.free[size]))
	heap.free[size] = append(heap.free[size], freeElt(ei))
}

var poison heapElt = heapElt{
	offset: MaxIndex,
	free:   MaxIndex,
	next:   MaxIndex,
	prev:   MaxIndex,
}

func (heap *Heap) removeFreeElt(ei, size Index) {
	e := &heap.elts[ei]
	fi := e.free
	if size >= Index(len(heap.free)) {
		size = 0
	}
	if l := Index(len(heap.free[size])); fi < l && heap.free[size][fi] == freeElt(ei) {
		if fi < l-1 {
			gi := heap.free[size][l-1]
			heap.free[size][fi] = gi
			heap.elts[gi].free = fi
		}
		heap.free[size] = heap.free[size][:l-1]
		*e = poison
		heap.removed = append(heap.removed, ei)
		return
	}
	panic("corrupt free list")
}

func (heap *Heap) eltSize(e *heapElt) Index {
	o := Index(heap.len)
	if e.next != MaxIndex {
		o = heap.elts[e.next].offset
	}
	return o - e.offset
}

func (heap *Heap) size(ei Index) Index { return heap.eltSize(&heap.elts[ei]) }

func (heap *Heap) Len(ei Index) uint {
	return uint(heap.size(ei))
}

func (heap *Heap) GetID(ei Index) (offset, len int) {
	e := &heap.elts[ei]
	return int(e.offset), int(heap.eltSize(e))
}

// Recycle previously removed elts.
func (heap *Heap) newElt() (ei Index) {
	if l := len(heap.removed); l > 0 {
		ei = heap.removed[l-1]
		heap.removed = heap.removed[:l-1]
		heap.elts[ei] = poison
	} else {
		ei = Index(len(heap.elts))
		heap.elts = append(heap.elts, poison)
	}
	return
}

func (heap *Heap) Get(size uint) (id Index, offset uint) { return heap.get(size, Index(size)) }

func (heap *Heap) get(sizeArg uint, size Index) (id Index, offset uint) {
	// Keep track of largest size caller asks for.
	if Index(sizeArg) > heap.maxSize {
		heap.maxSize = Index(sizeArg)
	}

	if size <= 0 {
		panic("size")
	}

	// Quickly allocate from free list of given size.
	if int(size) < len(heap.free) {
		if l := len(heap.free[size]); l > 0 {
			ei := heap.free[size][l-1]
			e := &heap.elts[ei]
			heap.free[size] = heap.free[size][:l-1]
			e.free = MaxIndex
			offset = uint(e.offset)
			id = Index(ei)
			return
		}
	}

	// Search free list 0: where free objects > max requested size are kept.
	if len(heap.free) > 0 {
		l := Index(len(heap.free[0]))
		for fi := Index(0); fi < l; fi++ {
			ei := heap.free[0][fi]
			e := &heap.elts[ei]
			es := heap.eltSize(e)
			fs := int(es) - int(size)
			if fs < 0 {
				continue
			}
			if fi < l-1 {
				gi := heap.free[0][l-1]
				heap.free[0][fi] = gi
				heap.elts[gi].free = fi
			}
			heap.free[0] = heap.free[0][:l-1]

			offset = uint(e.offset)
			e.free = MaxIndex
			id = Index(ei)

			if fs > 0 {
				heap.freeAfter(Index(ei), es, Index(fs))
			}
			return
		}
	}

	if heap.maxLen != 0 && heap.len+size > heap.maxLen {
		panic(fmt.Errorf("heap overflow allocating object of length %d", size))
	}

	if heap.len == 0 {
		heap.head = 0
		heap.tail = MaxIndex
	}

	ei := heap.newElt()
	e := &heap.elts[ei]

	offset = uint(heap.len)
	heap.len += size
	e.offset = Index(offset)

	e.next = MaxIndex
	e.prev = heap.tail
	e.free = MaxIndex

	heap.tail = ei

	if e.prev != MaxIndex {
		heap.elts[e.prev].next = ei
	}

	id = ei
	return
}

func (heap *Heap) newEltBefore(ei Index) (pi Index) {
	pi = heap.newElt()
	e, p := &heap.elts[ei], &heap.elts[pi]
	p.next = ei
	p.prev = e.prev
	if p.prev == MaxIndex {
		heap.head = pi
	} else {
		heap.elts[p.prev].next = pi
	}
	e.prev = pi
	return
}

func (heap *Heap) newEltAfter(ei Index) (ni Index) {
	ni = heap.newElt()
	e, n := &heap.elts[ei], &heap.elts[ni]
	n.prev = ei
	n.next = e.next
	if n.next == MaxIndex {
		heap.tail = ni
	} else {
		heap.elts[n.next].prev = ni
	}
	e.next = ni
	return
}

func (heap *Heap) GetAligned(sizeArg, log2Alignment uint) (id Index, offset uint) {
	// Adjust size for alignment so we guarantee a large enough block.
	a := Index(1) << log2Alignment

	size := sizeArg + uint(a) - 1
	sa := Index(sizeArg)
	s := Index(size)

	ei, offset := heap.get(sizeArg, s)
	o := Index(offset)

	// Aligned offset.
	ao := (o + a - 1) &^ (a - 1)

	if log2Alignment > 0 {
		if d := ao - o; d != 0 {
			pi := heap.newEltBefore(ei)
			e, p := &heap.elts[ei], &heap.elts[pi]
			p.offset = o
			e.offset = ao
			heap.Put(pi)
		}
		if d := int(o+s) - int(ao+sa); d > 0 {
			ni := heap.newEltAfter(ei)
			e, n := &heap.elts[ei], &heap.elts[ni]
			e.offset = ao
			n.offset = ao + sa
			heap.Put(ni)
		}
	}

	id = ei
	offset = uint(ao)
	return
}

func (heap *Heap) Put(ei Index) {
	e := &heap.elts[ei]

	if e.isFree() {
		panic(fmt.Errorf("duplicate free %d", ei))
	}

	// If previous element is free combine free elements.
	if e.prev != MaxIndex {
		prev := &heap.elts[e.prev]
		if prev.isFree() {
			ps := e.offset - prev.offset
			e.offset = prev.offset
			pi := e.prev
			e.prev = prev.prev
			if e.prev != MaxIndex {
				heap.elts[e.prev].next = ei
			}
			heap.removeFreeElt(pi, ps)
			if pi == heap.head {
				heap.head = ei
			}
		}
	}

	// If next element is free also combine.
	if e.next != MaxIndex {
		next := &heap.elts[e.next]
		if next.isFree() {
			ni := e.next
			ns := heap.size(ni)
			e.next = next.next
			if e.next != MaxIndex {
				heap.elts[e.next].prev = ei
			}
			heap.removeFreeElt(ni, ns)
			if ni == heap.tail {
				heap.tail = ei
			}
		}
	}

	es := heap.size(ei)
	heap.freeElt(ei, es)
}

func (heap *Heap) String() (s string) {
	s = fmt.Sprintf("%d elts", len(heap.elts))
	if heap.maxLen != 0 {
		s += fmt.Sprintf(", max %d elts (0x%x)", heap.maxLen, heap.maxLen)
	}
	return
}
