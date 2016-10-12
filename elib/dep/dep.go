package dep

import (
	"github.com/platinasystems/go/elib"
	"sort"
)

type Dep struct {
	Deps, AntiDeps []*Dep

	// User ordering.  We sort by increasing value of Order.
	Order uint32

	// Index into elts saved *before* sorting.
	index uint32

	// Bitmap of dependent indices considering both dependencies and anti-dependencies.
	depIndices elib.Bitmap
}

// Sort hooks by priority.
type Deps struct {
	elts      []*Dep
	isOrdered bool
	order     []uint32
}

func (hs *Deps) Add(h *Dep) {
	h.index = uint32(len(hs.elts))
	hs.elts = append(hs.elts, h)
	hs.isOrdered = false
}

func (h *Deps) Len() int      { return len(h.elts) }
func (h *Deps) Swap(i, j int) { h.elts[i], h.elts[j] = h.elts[j], h.elts[i] }
func (h *Deps) Less(i, j int) bool {
	// Respect user's ordering.
	if h.elts[i].Order != h.elts[j].Order {
		return h.elts[i].Order < h.elts[j].Order
	}
	// Otherwise maintain array order.
	return i > j
}

func (hs *Deps) orderHelper(h *Dep, b elib.Bitmap) elib.Bitmap {
	hi := uint(h.index)
	if !b.Get(hi) {
		b = b.Set(hi)
		for i := range h.Deps {
			b = hs.orderHelper(h.Deps[i], b)
		}
		hs.order = append(hs.order, h.index)
	}
	return b
}

func (d *Deps) index(i int, isForward bool) int {
	// No dependencies?  Use slice order.
	if d.elts == nil {
		return i
	}
	if !d.isOrdered {
		d.isOrdered = true
		d.sort()
	}
	if isForward {
		return int(d.order[i])
	} else {
		return int(d.order[len(d.order)-1-i])
	}
}

func (d *Deps) Index(i int) int        { return d.index(i, true) }
func (d *Deps) IndexReverse(i int) int { return d.index(i, false) }

// Indicate dependency.
func (h0 *Dep) dependsOn(h1 *Dep) { h0.depIndices = h0.depIndices.Set(uint(h1.index)) }

func (hs *Deps) sort() {
	// Indicate dependencies and anti-dependencies.
	for i := range hs.elts {
		for j := range hs.elts[i].Deps {
			hs.elts[i].dependsOn(hs.elts[i].Deps[j])
		}
		for j := range hs.elts[i].AntiDeps {
			hs.elts[i].AntiDeps[j].dependsOn(hs.elts[i])
		}
	}

	// Convert dependency bitmap to slice.
	for i := range hs.elts {
		h := hs.elts[i]
		if h.Deps != nil {
			h.Deps = h.Deps[:0]
		}
		j := ^uint(0)
		for h.depIndices.Next(&j) {
			h.Deps = append(h.Deps, hs.elts[j])
		}
		h.depIndices = h.depIndices.Free()
	}

	// Sort and form ordering.
	sort.Sort(hs)

	if hs.order != nil {
		hs.order = hs.order[:0]
	}
	v := elib.Bitmap(0)
	for i := range hs.elts {
		v = hs.orderHelper(hs.elts[i], v)
	}
	v = v.Free()
}
