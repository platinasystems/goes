// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/event"
	"github.com/platinasystems/go/elib/loop"
)

type eventNode struct{ Node }

type eventMain struct {
	eventNode eventNode
}

func (v *Vnet) eventInit() {
	n := &v.eventMain.eventNode
	n.Vnet = v
	v.loop.RegisterNode(n, "vnet")
}
func (v *Vnet) CurrentEvent() *loop.Event { return v.eventNode.CurrentEvent() }

type Event struct {
	loop.Event
	n *Node
}

func (e *Event) Node() *Node      { return e.n }
func (e *Event) Vnet() *Vnet      { return e.n.Vnet }
func (e *Event) GetEvent() *Event { return e }

type Eventer interface {
	GetEvent() *Event
	event.Actor
}

func (n *Node) SignalEventp(r Eventer, p elog.PointerToFirstArg) {
	v := n.Vnet
	e := r.GetEvent()
	e.n = n
	n.AddEventp(r, &v.eventMain.eventNode, p)
}
func (n *Node) SignalEvent(r Eventer) { n.SignalEventp(r, elog.PointerToFirstArg(&n)) }

func (n *Node) AddTimedEventp(r Eventer, dt float64, p elog.PointerToFirstArg) {
	v := n.Vnet
	e := r.GetEvent()
	e.n = n
	n.Node.AddTimedEventp(r, &v.eventMain.eventNode, dt, p)
}
func (n *Node) AddTimedEvent(r Eventer, dt float64) {
	n.AddTimedEventp(r, dt, elog.PointerToFirstArg(&n))
}

func (e *Event) Signal(r Eventer) { e.n.SignalEventp(r, elog.PointerToFirstArg(&e)) }
func (e *Event) AddTimedEvent(r Eventer, dt float64) {
	e.n.AddTimedEventp(r, dt, elog.PointerToFirstArg(&e))
}
func (v *Vnet) SignalEvent(r Eventer) { v.eventNode.SignalEventp(r, elog.PointerToFirstArg(&v)) }
func (v *Vnet) AddTimedEvent(r Eventer, dt float64) {
	v.eventNode.AddTimedEventp(r, dt, elog.PointerToFirstArg(&v))
}
