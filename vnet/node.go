// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/dep"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/elib/loop"

	"fmt"
)

type Node struct {
	Vnet *Vnet
	loop.Node
	Dep       dep.Dep
	Errors    []string
	errorRefs []ErrorRef
}

func (n *Node) GetVnetNode() *Node { return n }

func (n *Node) AddSuspendActivity(i *RefIn, a int) {
	n.Vnet.loop.AddSuspendActivity(&i.In, a, &suspendLimits)
}
func (n *Node) Suspend(i *RefIn) { n.Vnet.loop.Suspend(&i.In, &suspendLimits) }
func (n *Node) Resume(i *RefIn)  { n.Vnet.loop.Resume(&i.In, &suspendLimits) }

const MaxVectorLen = loop.MaxVectorLen

type Noder interface {
	loop.Noder
	GetVnetNode() *Node
}

func (v *Vnet) AddNamedNext(n Noder, name string) uint {
	if nextIndex, err := v.loop.AddNamedNext(n, name); err == nil {
		return nextIndex
	} else {
		panic(err)
	}
}

type InputNode struct {
	Node
	o InputNoder
}

func (n *InputNode) GetInputNode() *InputNode                 { return n }
func (n *InputNode) MakeLoopOut() loop.LooperOut              { return &RefOut{} }
func (n *InputNode) LoopInput(l *loop.Loop, o loop.LooperOut) { n.o.NodeInput(o.(*RefOut)) }

type InputNoder interface {
	Noder
	GetInputNode() *InputNode
	NodeInput(o *RefOut)
}

func (v *Vnet) RegisterInputNode(n InputNoder, name string, args ...interface{}) {
	v.RegisterNode(n, name, args...)
	x := n.GetInputNode()
	x.o = n
}

type OutputNode struct {
	Node
	o OutputNoder
}

func (n *OutputNode) GetOutputNode() *OutputNode               { return n }
func (n *OutputNode) MakeLoopIn() loop.LooperIn                { return &RefIn{} }
func (n *OutputNode) LoopOutput(l *loop.Loop, i loop.LooperIn) { n.o.NodeOutput(i.(*RefIn)) }

type OutputNoder interface {
	Noder
	GetOutputNode() *OutputNode
	NodeOutput(i *RefIn)
}

func (v *Vnet) RegisterOutputNode(n OutputNoder, name string, args ...interface{}) {
	v.RegisterNode(n, name, args...)
	x := n.GetOutputNode()
	x.o = n
}

type enqueue struct {
	x, n uint32
	v    *Vnet
	i    *RefIn
	o    *RefOut
}

func (q *enqueue) sync() {
	l := q.o.Outs[q.x].GetLen(q.v)
	if n := uint(q.n); n > l {
		q.o.Outs[q.x].Dup(q.i)
		q.o.Outs[q.x].SetLen(q.v, n)
	}
}
func (q *enqueue) validate() {
	if !elib.Debug {
		return
	}
	out_len, in_len := uint(0), q.i.InLen()
	for i := range q.o.Outs {
		o := &q.o.Outs[i]
		out_len += o.GetLen(q.v)
	}
	if out_len > in_len {
		panic(fmt.Errorf("out len %d > in len %d", out_len, in_len))
	}
}

func (q *enqueue) put(r0 *Ref, x0 uint) {
	q.o.Outs[x0].Dup(q.i)
	i0 := q.o.Outs[x0].AddLen(q.v)
	q.o.Outs[x0].Refs[i0] = *r0
}
func (q *enqueue) Put1(r0 *Ref, x0 uint) {
	q.o.Outs[q.x].Refs[q.n] = *r0
	q.n++
	if uint32(x0) != q.x {
		q.n--
		q.put(r0, x0)
	}
}

func (q *enqueue) setCachedNext(x0 uint) {
	q.sync()
	// New cached next and count.
	q.x = uint32(x0)
	q.n = uint32(q.o.Outs[x0].GetLen(q.v))
}

func (q *enqueue) Put2(r0, r1 *Ref, x0, x1 uint) {
	// Speculatively enqueue both refs to cached next.
	n0 := q.n
	q.o.Outs[q.x].Refs[n0+0] = *r0 // (*) see below.
	q.o.Outs[q.x].Refs[n0+1] = *r1
	q.n = n0 + 2

	// Confirm speculation.
	same, match_cache0 := x0 == x1, uint32(x0) == q.x
	if same && match_cache0 {
		return
	}

	// Restore cached length.
	q.n = n0

	// Put refs in correct next slots.
	q.Put1(r0, x0)
	q.Put1(r1, x1)

	// If neither next matches cached next and both are the same, then changed cached next.
	if same {
		q.setCachedNext(x0)
	}
}

func (q *enqueue) Put4(r0, r1, r2, r3 *Ref, x0, x1, x2, x3 uint) {
	// Speculatively enqueue both refs to cached next.
	n0, x := q.n, uint(q.x)
	q.o.Outs[x].Refs[n0+0] = *r0
	q.o.Outs[x].Refs[n0+1] = *r1
	q.o.Outs[x].Refs[n0+2] = *r2
	q.o.Outs[x].Refs[n0+3] = *r3
	q.n = n0 + 4

	// Confirm speculation.
	if x0 == x && x0 == x1 && x2 == x3 && x0 == x2 {
		return
	}

	// Restore cached length.
	q.n = n0

	// Put refs in correct next slots.
	q.Put1(r0, x0)
	q.Put1(r1, x1)
	q.Put1(r2, x2)
	q.Put1(r3, x3)

	// If last 2 misses in cache and both are the same, then changed cached next.
	if x2 != x && x2 == x3 {
		q.setCachedNext(x2)
	}
}

//go:generate gentemplate -d Package=vnet -id enqueue -d VecType=enqueue_vec -d Type=*enqueue github.com/platinasystems/go/elib/vec.tmpl

type InOutNode struct {
	Node
	qs enqueue_vec
	t  InOutNoder
}

func (n *InOutNode) GetEnqueue(in *RefIn) (q *enqueue) {
	i := in.ThreadId()
	n.qs.Validate(i)
	q = n.qs[i]
	if n.qs[i] == nil {
		q = &enqueue{}
		n.qs[i] = q
	}
	return
}

func (n *InOutNode) GetInOutNode() *InOutNode    { return n }
func (n *InOutNode) MakeLoopIn() loop.LooperIn   { return &RefIn{} }
func (n *InOutNode) MakeLoopOut() loop.LooperOut { return &RefOut{} }
func (n *InOutNode) LoopInputOutput(l *loop.Loop, i loop.LooperIn, o loop.LooperOut) {
	in, out := i.(*RefIn), o.(*RefOut)
	q := n.GetEnqueue(in)
	q.n, q.i, q.o, q.v = 0, in, out, n.Vnet
	n.t.NodeInput(in, out)
	q.sync()
	q.validate()
}

type InOutNoder interface {
	Noder
	GetInOutNode() *InOutNode
	NodeInput(i *RefIn, o *RefOut)
}

func (v *Vnet) RegisterInOutNode(n InOutNoder, name string, args ...interface{}) {
	v.RegisterNode(n, name, args...)
	x := n.GetInOutNode()
	x.t = n
}

// Main structure.
type Vnet struct {
	loop loop.Loop
	hw.BufferMain
	cliMain cliMain
	eventMain
	interfaceMain
	packageMain
}

func (v *Vnet) GetLoop() *loop.Loop { return &v.loop }

func (v *Vnet) RegisterNode(n Noder, format string, args ...interface{}) {
	v.loop.RegisterNode(n, format, args...)
	x := n.GetVnetNode()
	x.Vnet = v

	x.errorRefs = make([]ErrorRef, len(x.Errors))
	for i := range x.Errors {
		er := ^ErrorRef(0)
		if len(x.Errors[i]) > 0 {
			er = x.NewError(x.Errors[i])
		}
		x.errorRefs[i] = er
	}
}

func (node *Node) Redirect(in *RefIn, out *RefOut, next uint) {
	o := &out.Outs[next]
	n := in.InLen()
	copy(o.Refs[:n], in.Refs[:n])
	node.SetOutLen(o, in, n)
}

func (node *Node) ErrorRedirect(in *RefIn, out *RefOut, next uint, err ErrorRef) {
	o := &out.Outs[next]
	n := in.InLen()
	for i := uint(0); i < n; i++ {
		r := &o.Refs[i]
		*r = in.Refs[i]
		r.Aux = uint32(err)
	}
	node.SetOutLen(o, in, n)
}
