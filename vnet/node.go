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

func (q *enqueue) put(x0 uint, r0 *Ref) {
	q.o.Outs[x0].Dup(q.i)
	i0 := q.o.Outs[x0].AddLen(q.v)
	q.o.Outs[x0].Refs[i0] = *r0
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

func (q *enqueue) Put1(r0 *Ref, x0 uint) {
	q.o.Outs[q.x].Refs[q.n] = *r0
	q.n++
	if uint32(x0) != q.x {
		q.n--
		q.put(x0, r0)
	}
}

func (q *enqueue) Put2(r0, r1 *Ref, x0, x1 uint) {
	n0 := q.n
	q.o.Outs[q.x].Refs[n0+0] = *r0
	q.o.Outs[q.x].Refs[n0+1] = *r1
	q.n = n0 + 2
	if same := x0 == x1; !same || uint32(x0) != q.x {
		q.n = n0
		q.sync()
		q.put(x0, r0)
		q.put(x1, r1)
		if same {
			q.x = uint32(x0)
			q.n = uint32(q.o.Outs[x0].GetLen(q.v))
		}
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
