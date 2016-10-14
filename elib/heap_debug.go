// +build debug

package elib

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

type visited uint8

const (
	unvisited visited = iota
	visitedAlloc
	visitedFree
	removed
)

var visitedElts []visited

func (p *Heap) validate() (err error) {
	if cap(visitedElts) < len(p.elts) {
		visitedElts = make([]visited, len(p.elts))
	}
	nVisited := 0
	visitedElts = visitedElts[:len(p.elts)]
	for i := range visitedElts {
		visitedElts[i] = unvisited
	}

	ei := p.head
	prev := MaxIndex
	prevIsFree := false
	offset := Index(0)
	for ei != MaxIndex {
		e := &p.elts[ei]

		if visitedElts[ei] != unvisited {
			err = fmt.Errorf("duplicate visit %+v", e)
			return
		}

		if visitedElts[ei] != unvisited {
			err = fmt.Errorf("already visited %+v", e)
			return
		}

		if e.prev != prev {
			err = fmt.Errorf("bad prev pointer %+v", e)
			return
		}

		if e.offset != offset {
			err = fmt.Errorf("bad offset %d %+v", offset, e)
			return
		}

		size := p.size(ei)
		if size == 0 {
			err = fmt.Errorf("zero size %d %+v", ei, e)
			return
		}
		offset += size

		visitedElts[ei] = visitedAlloc
		nVisited++

		isFree := e.isFree()
		if isFree {
			visitedElts[ei] = visitedFree
			if size > p.maxSize {
				size = 0
			}
			if prevIsFree {
				err = fmt.Errorf("two consecutive free elts")
				return
			}
			if size >= Index(len(p.free)) {
				err = fmt.Errorf("size %d >= len free %d", size, len(p.free))
				return
			}
			if e.free >= Index(len(p.free[size])) {
				err = fmt.Errorf("free %d >= len free[%d] %d, size %d %+v", e.free, size, len(p.free[size]), p.size(ei), e)
				return
			}
			if ei != Index(p.free[size][e.free]) {
				err = fmt.Errorf("corrupt free list %d != free[%d][%d] %d", ei, size, e.free, p.free[size][e.free])
				return
			}
		}

		prev = ei
		prevIsFree = isFree
		ei = e.next
	}
	if prev != p.tail {
		err = fmt.Errorf("corrupt tail %d != %d", prev, p.tail)
		return
	}

	for i := range p.removed {
		ei := p.removed[i]
		e := &p.elts[ei]
		if visitedElts[ei] != unvisited {
			err = fmt.Errorf("removed visited %+v", e)
			return
		}
		visitedElts[ei] = removed
		nVisited++
	}

	for ei := range p.elts {
		if visitedElts[ei] == unvisited {
			err = fmt.Errorf("unvisited elt %d", ei)
			return
		}
	}

	// Make sure all elts have been visited
	if nVisited != len(visitedElts) {
		err = fmt.Errorf("visited %d != %d", nVisited, len(visitedElts))
		return
	}

	return
}

type testHeap struct {
	heap Heap

	// Number of iterations to run
	iterations Count

	// Validate/print every so many iterations (zero means never).
	validateEvery Count
	printEvery    Count

	// Seed to make randomness deterministic.  0 means choose seed.
	seed int64

	// Number of objects to create.
	nObjects Count

	// log2 max size of objects.
	log2MaxLen int

	// Random max alignment of objects.
	maxAlign int

	verbose int
}

func HeapTest() {
	t := testHeap{
		iterations: 10,
		nObjects:   10,
		verbose:    1,
	}
	flag.Var(&t.iterations, "iter", "Number of iterations")
	flag.Var(&t.validateEvery, "valid", "Number of iterations per validate")
	flag.Var(&t.printEvery, "print", "Number of iterations per print")
	flag.Int64Var(&t.seed, "seed", 0, "Seed for random number generator")
	flag.IntVar(&t.log2MaxLen, "len", 8, "Log2 max length of object to allocate")
	flag.Var(&t.nObjects, "objects", "Number of random objects")
	flag.IntVar(&t.maxAlign, "align", 0, "Maximum random alignment of objects")
	flag.IntVar(&t.verbose, "verbose", 0, "Be verbose")
	flag.Parse()

	err := runHeapTest(&t)
	if err != nil {
		panic(err)
	}
}

type randHeapObj struct {
	id     Index
	len    uint
	offset uint
	align  uint
}

//go:generate gentemplate -d Package=elib -id randHeapObj -d VecType=randHeapObjVec -tags debug -d Type=randHeapObj vec.tmpl

func (t *testHeap) validate(iter int) (err error) {
	if t.validateEvery != 0 && iter%int(t.validateEvery) == 0 {
		if err = t.heap.validate(); err != nil {
			if t.verbose != 0 {
				fmt.Printf("iter %d: %s\n%+v\n", iter, err, &t.heap)
			}
			return
		}
	}
	if t.printEvery != 0 && iter%int(t.printEvery) == 0 {
		fmt.Printf("%10g iterations: %s\n", float64(iter), &t.heap)
	}
	return
}

func runHeapTest(t *testHeap) (err error) {
	var p *Heap = &t.heap
	var s Uint64Vec
	var objs randHeapObjVec

	if t.seed == 0 {
		t.seed = int64(time.Now().Nanosecond())
	}

	rand.Seed(t.seed)
	if t.verbose != 0 {
		fmt.Printf("%#v\n", t)
	}
	objs.Resize(uint(t.nObjects))
	var iter int

	for iter = 0; iter < int(t.iterations); iter++ {
		o := &objs[rand.Int()%len(objs)]
		if o.len != 0 {
			if l := p.Len(o.id); l != o.len {
				err = fmt.Errorf("len mismatch %d != %d", l, o.len)
				return
			}
			p.Put(o.id)
			o.len = 0
		} else {
			o.len = 1 + uint(rand.Int()&(1<<uint(t.log2MaxLen)-1))
			o.align = 0
			if t.maxAlign > 0 {
				a := uint(t.maxAlign)
				for 1<<a > o.len {
					a--
				}
				if a > 0 {
					o.align = uint(rand.Intn(int(a)))
				}
			}
			o.id, o.offset = p.GetAligned(o.len, o.align)
			if o.offset&((1<<o.align)-1) != 0 {
				panic("unaligned")
			}
			s.Validate(o.offset + o.len - 1)
			for j := uint(0); j < o.len; j++ {
				s[o.offset+j] = uint64(o.id)<<uint(t.log2MaxLen) + uint64(o.offset+j)
			}
		}
		err = t.validate(iter)
		if err != nil {
			return
		}
	}
	if t.verbose != 0 {
		fmt.Printf("%d iterations: %+v\n", iter, p)
	}
	for i := range objs {
		o := &objs[i]
		if o.len > 0 {
			p.Put(o.id)
		}
		err = t.validate(iter)
		if err != nil {
			return
		}
		iter++
	}
	if t.verbose != 0 {
		fmt.Printf("%d iterations: %+v\n", iter, p)
		fmt.Printf("No errors: %d iterations\n", t.iterations)
	}
	return
}
