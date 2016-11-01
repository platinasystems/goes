// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"github.com/platinasystems/go/elib/event"
)

type eventNode struct{ Node }

func (n *eventNode) EventHandler() {}

type eventMain struct {
	eventNode eventNode
}

func (v *Vnet) eventInit() {
	n := &v.eventMain.eventNode
	n.Vnet = v
	v.loop.RegisterNode(n, "event-handler")
}

type Event struct {
	n *Node
}

func (e *Event) Node() *Node      { return e.n }
func (e *Event) Vnet() *Vnet      { return e.n.Vnet }
func (e *Event) GetEvent() *Event { return e }

type Eventer interface {
	GetEvent() *Event
	event.Actor
}

func (n *Node) SignalEvent(r Eventer) {
	v := n.Vnet
	e := r.GetEvent()
	e.n = n
	n.AddEvent(r, &v.eventMain.eventNode)
}

func (n *Node) AddTimedEvent(r Eventer, dt float64) {
	v := n.Vnet
	e := r.GetEvent()
	e.n = n
	n.Node.AddTimedEvent(r, &v.eventMain.eventNode, dt)
}

func (e *Event) Signal(r Eventer)                    { e.n.SignalEvent(r) }
func (e *Event) AddTimedEvent(r Eventer, dt float64) { e.n.AddTimedEvent(r, dt) }
func (v *Vnet) SignalEvent(r Eventer)                { v.eventNode.SignalEvent(r) }
func (v *Vnet) AddTimedEvent(r Eventer, dt float64)  { v.eventNode.AddTimedEvent(r, dt) }
