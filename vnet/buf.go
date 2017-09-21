// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/elib/loop"

	"fmt"
	"reflect"
	"sort"
	"unsafe"
)

type RefOpaque struct {
	// Software interface.
	Si Si

	// Aux data.
	// For example, used by error node to give reason for dropping this packet.
	Aux uint32
}

type Ref struct {
	hw.RefHeader
	RefOpaque
}

func (r *Ref) Flags() BufferFlag         { return BufferFlag(r.RefHeader.Flags()) }
func (r *Ref) NextValidFlag() BufferFlag { return BufferFlag(r.RefHeader.NextValidFlag()) }
func (r *Ref) NextRef() *Ref {
	return (*Ref)(unsafe.Pointer(r.RefHeader.NextRef()))
}

type BufferState hw.BufferState

const (
	BufferUnknown        = BufferState(hw.BufferUnknown)
	BufferKnownAllocated = BufferState(hw.BufferKnownAllocated)
	BufferKnownFree      = BufferState(hw.BufferKnownFree)
)

func (r *Ref) ValidateState(p *BufferPool, s BufferState) {
	r.RefHeader.ValidateState((*hw.BufferPool)(p), hw.BufferState(s))
}
func (r *Ref) Trace(p *BufferPool, i hw.BufferTracer, e int) {
	r.RefHeader.Trace((*hw.BufferPool)(p), i, e)
}

func (r *Ref) Foreach(f func(r *Ref, i uint)) {
	i := uint(0)
	for {
		f(r, i)
		if r = r.NextRef(); r == nil {
			break
		}
		i++
	}
}

type BufferFlag hw.BufferFlag

const (
	NextValid = BufferFlag(hw.NextValid)
	Cloned    = BufferFlag(hw.Cloned)
)

func RefFlag1(f BufferFlag, r []Ref, i uint) bool {
	return hw.RefFlag1(hw.BufferFlag(f), &r[i+0].RefHeader)
}
func RefFlag2(f BufferFlag, r []Ref, i uint) bool {
	return hw.RefFlag2(hw.BufferFlag(f), &r[i+0].RefHeader, &r[i+1].RefHeader)
}
func RefFlag4(f BufferFlag, r []Ref, i uint) bool {
	return hw.RefFlag4(hw.BufferFlag(f), &r[i+0].RefHeader, &r[i+1].RefHeader, &r[i+2].RefHeader, &r[i+3].RefHeader)
}

type RefChain hw.RefChain

func (c *RefChain) Len() uint  { return (*hw.RefChain)(c).Len() }
func (c *RefChain) Reset()     { *c = RefChain{} }
func (c *RefChain) Head() *Ref { return (*Ref)(unsafe.Pointer((*hw.RefChain)(c).Head())) }
func (c *RefChain) Validate()  { (*hw.RefChain)(c).Validate() }

func (c *RefChain) Append(r *Ref) {
	if c.Len() == 0 {
		h := c.Head()
		*h = *r
	}
	(*hw.RefChain)(c).Append(&r.RefHeader)
	c.Validate()
}
func (c *RefChain) Done() (h Ref) {
	h = *c.Head()
	c.Validate()
	c.Reset()
	return
}

//go:generate gentemplate -d Package=vnet -id Ref -d VecType=RefVec -d Type=Ref github.com/platinasystems/go/elib/vec.tmpl

type refInCommon struct {
	loop.In
	refInDupCopy
}

type refInDupCopy struct {
	BufferPool *BufferPool
}

func (i *RefIn) Dup(x *RefIn) { i.refInDupCopy = x.refInDupCopy }

type RefIn struct {
	refInCommon
	Refs [MaxVectorLen]Ref
}

func (r *RefIn) Cap() uint { return uint(len(r.Refs)) }

type RefVecIn struct {
	refInCommon

	// Number of packets corresponding to vector of buffer refs.
	nPackets uint

	Refs RefVec
}

type RefOut struct {
	loop.Out
	Outs []RefIn
}

type BufferPool hw.BufferPool

var DefaultBufferPool = &BufferPool{
	Name:           "default",
	BufferTemplate: getDefaultBufferTemplate(),
}

func getDefaultBufferTemplate() hw.BufferTemplate {
	t := *hw.DefaultBufferTemplate
	r := (*Ref)(unsafe.Pointer(&t.Ref))
	// Poison so that if user does not override its obvious.
	r.Aux = poisonErrorRef
	r.Si = ^r.Si
	return t
}

func (p *BufferPool) GetRefTemplate() *Ref {
	return (*Ref)(unsafe.Pointer(&p.BufferTemplate.Ref))
}
func (p *BufferPool) GetBufferTemplate() *Buffer {
	return (*Buffer)(unsafe.Pointer(&p.BufferTemplate.Buffer))
}

func (v *Vnet) AddBufferPool(p *BufferPool) {
	v.BufferMain.AddBufferPool((*hw.BufferPool)(p))
}

func (r *RefIn) AllocPoolRefs(p *BufferPool, n uint) {
	r.BufferPool = p
	(*hw.BufferPool)(p).AllocRefs(&r.Refs[0].RefHeader, n)
}
func (r *RefIn) FreePoolRefs(p *BufferPool, n uint) {
	const freeNext = true
	(*hw.BufferPool)(p).FreeRefs(&r.Refs[0].RefHeader, n, freeNext)
}

func (p *BufferPool) AllocCachedRefs() (r RefVec) {
	rs := (*hw.BufferPool)(p).AllocCachedRefs()
	if len(rs) > 0 {
		r = (*RefHeader)(&rs[0].RefHeader).slice(rs.Len())
	}
	return
}

func (p *BufferPool) AllocRefsStride(r *Ref, n, stride uint) {
	(*hw.BufferPool)(p).AllocRefsStride((*hw.RefHeader)(&r.RefHeader), n, stride)
}
func (p *BufferPool) AllocRefs(r RefVec) { p.AllocRefsStride(&r[0], r.Len(), 1) }
func (p *BufferPool) FreeRefs(r *Ref, n uint, freeNext bool) {
	(*hw.BufferPool)(p).FreeRefs((*hw.RefHeader)(&r.RefHeader), n, freeNext)
}

func (i *RefIn) AllocRefs(n uint) { i.AllocPoolRefs(i.BufferPool, n) }
func (i *RefIn) FreeRefs(n uint)  { i.FreePoolRefs(i.BufferPool, n) }

func (in *RefIn) SetLen(v *Vnet, new_len uint) {
	if elib.Debug {
		old_len := in.In.GetLen(&v.loop)
		for i := old_len; i < new_len; i++ {
			in.BufferPool.ValidateRef(&in.Refs[i], BufferKnownAllocated)
		}
	}
	in.In.SetLen(&v.loop, new_len)
}

func (i *RefIn) GetLen(v *Vnet) uint { return i.In.GetLen(&v.loop) }

func (i *RefIn) AddLen(v *Vnet) (l uint) {
	l = i.In.GetLen(&v.loop)
	i.In.SetLen(&v.loop, l+1)
	return
}
func (i *RefIn) SetPoolAndLen(v *Vnet, p *BufferPool, l uint) {
	i.BufferPool = p
	i.SetLen(v, l)
}

func Get4Refs(rs []Ref, i uint) (r0, r1, r2, r3 *Ref) {
	r0, r1, r2, r3 = &rs[i+0], &rs[i+1], &rs[i+2], &rs[i+3]
	return
}

func (p *BufferPool) ValidateRef(r *Ref, want BufferState) {
	(*hw.BufferPool)(p).ValidateRefs((*hw.RefHeader)(&r.RefHeader), hw.BufferState(want), 1, 1)
}

func (p *BufferPool) ValidateRefs(refs []Ref, want hw.BufferState) {
	(*hw.BufferPool)(p).ValidateRefs((*hw.RefHeader)(&refs[0].RefHeader), want, uint(len(refs)), 1)
}

type Buffer hw.Buffer

func (r *Ref) GetBuffer() *Buffer { return (*Buffer)(r.RefHeader.GetBuffer()) }

func Get4Buffers(rs []Ref, i uint) (b0, b1, b2, b3 *Buffer) {
	b0, b1, b2, b3 = rs[i+0].GetBuffer(), rs[i+1].GetBuffer(), rs[i+2].GetBuffer(), rs[i+3].GetBuffer()
	return
}

func Get1Buffer(rs []Ref, i uint) (b0 *Buffer) {
	b0 = rs[i+0].GetBuffer()
	return
}

func (r *RefIn) Get1(i uint) (_ *Ref)    { return &r.Refs[i] }
func (r *RefIn) Get2(i uint) (_, _ *Ref) { return &r.Refs[i+0], &r.Refs[i+1] }
func (r *RefIn) Get4(i uint) (_, _, _, _ *Ref) {
	return &r.Refs[i+0], &r.Refs[i+1], &r.Refs[i+2], &r.Refs[i+3]
}

func (n *Node) SetOutLen(out *RefIn, in *RefIn, l uint) {
	out.Dup(in)
	out.SetLen(n.Vnet, l)
}

func (r *RefVecIn) FreePoolRefs(p *BufferPool, freeNext bool) {
	l := r.Refs.Len()
	if l > 0 {
		(*hw.BufferPool)(p).FreeRefs(&r.Refs[0].RefHeader, l, freeNext)
	}
}
func (r *RefVecIn) Len() uint              { return r.Refs.Len() }
func (r *RefVecIn) NPackets() uint         { return r.nPackets }
func (r *RefVecIn) FreeRefs(freeNext bool) { r.FreePoolRefs(r.BufferPool, freeNext) }

type RefHeader hw.RefHeader

func (r *RefHeader) slice(n uint) (l []Ref) {
	var h reflect.SliceHeader
	h.Data = uintptr(unsafe.Pointer(r))
	h.Len = int(n)
	h.Cap = int(n)
	l = *(*[]Ref)(unsafe.Pointer(&h))
	return
}

type showPool struct {
	Pool string `format:"%-30s" align:"left"`
	Size string `format:"%-12s" align:"right"`
	Free string `format:"%-12s" align:"right"`
	Used string `format:"%-12s" align:"right"`
}
type showPools []showPool

func (x showPools) Less(i, j int) bool { return x[i].Pool < x[j].Pool }
func (x showPools) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x showPools) Len() int           { return len(x) }

func (v *Vnet) showBufferUsage(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	m := &v.BufferMain

	sps := []showPool{}
	fmt.Fprintf(w, "DMA heap: %s\n", hw.DmaHeapUsage())
	for _, p := range m.PoolByName {
		sps = append(sps, showPool{
			Pool: p.Name,
			Size: fmt.Sprintf("%12d", p.Size),
			Free: fmt.Sprintf("%12s", elib.MemorySize(p.SizeIncludingOverhead()*p.FreeLen())),
			Used: fmt.Sprintf("%12s", elib.MemorySize(p.DmaMemAllocBytes)),
		})
	}
	sort.Sort(showPools(sps))
	elib.Tabulate(sps).Write(w)
	return
}
