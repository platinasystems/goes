// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"github.com/platinasystems/go/elib/dep"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/elib/loop"
)

type Node struct {
	Vnet *Vnet
	loop.Node
	Dep       dep.Dep
	Errors    []string
	errorRefs []ErrorRef
}

func (n *Node) GetVnetNode() *Node { return n }
func (n *Node) Suspend(i *RefIn)   { n.Vnet.loop.Suspend(&i.In) }
func (n *Node) Resume(i *RefIn)    { n.Vnet.loop.Resume(&i.In) }

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
	cached_next   uint32
	n_cached_next uint32
	v             *Vnet
	i             *RefIn
	o             *RefOut
}

func (q *enqueue) x(x0 uint, r0 *Ref) {
	q.o.Outs[x0].Dup(q.i)
	i0 := q.o.Outs[x0].AddLen(q.v)
	q.o.Outs[x0].Refs[i0] = *r0
}
func (q *enqueue) sync() {
	if q.n_cached_next > 0 {
		q.o.Outs[q.cached_next].SetLen(q.v, uint(q.n_cached_next))
		q.n_cached_next = 0
	}
}

func (q *enqueue) Put1(r0 *Ref, x0 uint) {
	q.o.Outs[q.cached_next].Refs[q.n_cached_next] = *r0
	q.n_cached_next++
	if uint32(x0) != q.cached_next {
		q.n_cached_next--
		q.x(x0, r0)
	}
}

func (q *enqueue) Put2(r0, r1 *Ref, x0, x1 uint) {
	q.o.Outs[q.cached_next+0].Refs[q.n_cached_next] = *r0
	q.o.Outs[q.cached_next+1].Refs[q.n_cached_next] = *r1
	q.n_cached_next += 2
	if same := x0 == x1; !same || uint32(x0) != q.cached_next {
		q.n_cached_next -= 2
		q.x(x0, r0)
		q.x(x1, r1)
		if same {
			q.sync()
			q.cached_next = uint32(x0)
		}
	}
}

type InOutNode struct {
	Node
	enqueue
	t InOutNoder
}

func (n *InOutNode) GetInOutNode() *InOutNode    { return n }
func (n *InOutNode) MakeLoopIn() loop.LooperIn   { return &RefIn{} }
func (n *InOutNode) MakeLoopOut() loop.LooperOut { return &RefOut{} }
func (n *InOutNode) LoopInputOutput(l *loop.Loop, i loop.LooperIn, o loop.LooperOut) {
	q, in, out := &n.enqueue, i.(*RefIn), o.(*RefOut)
	q.i, q.o, q.v = in, out, n.Vnet
	n.t.NodeInput(in, out)
	q.sync()
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
	cliMain
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
