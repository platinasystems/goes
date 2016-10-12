package ip

import (
	"github.com/platinasystems/go/vnet"
)

// Dense index into fib vector.
type FibIndex uint32

//go:generate gentemplate -d Package=ip -id FibIndex -d VecType=FibIndexVec -d Type=FibIndex github.com/platinasystems/go/elib/vec.tmpl

// Sparse 32 bit id for route table.
type FibId uint32

type fibMain struct {
	// Table index indexed by software interface.
	fibIndexBySi FibIndexVec

	// Hash table mapping table id to fib index.
	// ID space is not necessarily dense; index space is dense.
	fibIndexById map[FibId]FibIndex

	// Hash table mapping interface route rewrite adjacency index by sw if index.
	ifRouteAdjBySi map[vnet.Si]FibIndex
}

func (f *fibMain) fibIndexForSi(si vnet.Si, validate bool) FibIndex {
	if validate {
		f.fibIndexBySi.Validate(uint(si))
	}
	return f.fibIndexBySi[si]
}
func (f *fibMain) FibIndexForSi(si vnet.Si) FibIndex {
	return f.fibIndexForSi(si, false)
}
func (f *fibMain) ValidateFibIndexForSi(si vnet.Si) FibIndex {
	return f.fibIndexForSi(si, true)
}
func (f *fibMain) FibIndexForId(id FibId) (i FibIndex, ok bool) { i, ok = f.fibIndexById[id]; return }
func (f *fibMain) SetFibIndexForId(id FibId, i FibIndex) {
	if f.fibIndexById == nil {
		f.fibIndexById = make(map[FibId]FibIndex)
	}
	f.fibIndexById[id] = i
}
