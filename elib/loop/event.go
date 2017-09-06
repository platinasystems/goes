// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loop

import (
	"github.com/platinasystems/go/elib/cpu"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/event"

	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

type EventPoller interface {
	EventPoll()
}

type EventHandler interface {
	Noder
	EventHandler()
}

type eventMain struct {
	l        *Loop
	pollers  []EventPoller
	handlers []EventHandler

	loopEvents    sync.Pool
	events        chan *loopEvent
	eventPoolLock sync.Mutex
	eventPool     event.Pool
	eventVec      event.ActorVec
}

func (l *eventMain) getLoopEvent(a event.Actor, p elog.PointerToFirstArg) (x *loopEvent) {
	if y := l.loopEvents.Get(); y != nil {
		x = y.(*loopEvent)
		*x = loopEvent{actor: a}
	} else {
		x = &loopEvent{actor: a}
	}
	x.l = l.l
	x.caller = elog.GetCaller(p)
	return
}
func (l *eventMain) putLoopEvent(x *loopEvent) { l.loopEvents.Put(x) }

type eventNode struct {
	activateEvent
	sequence uint
	rxEvents chan *loopEvent
	eventWg  sync.WaitGroup
	eventVec event.ActorVec
}

type loopEvent struct {
	l      *Loop
	actor  event.Actor
	caller elog.Caller
	dst    *Node
	time   cpu.Time
}

func (e *loopEvent) EventTime() cpu.Time { return e.time }

func (l *Loop) addEvent(le *loopEvent) {
	select {
	case l.events <- le:
	default:
		l.addTimedEvent(le, 0)
	}
}

func (l *Loop) addTimedEvent(le *loopEvent, dt float64) {
	le.time = cpu.TimeNow() + cpu.Time(dt*l.cyclesPerSec)
	l.eventPoolLock.Lock()
	defer l.eventPoolLock.Unlock()
	l.eventPool.Add(le)
}

// AddEvent adds event whose action will be called on the next loop iteration.
func (n *Node) AddEventp(e event.Actor, dst EventHandler, p elog.PointerToFirstArg) {
	le := n.l.getLoopEvent(e, p)
	if dst != nil {
		le.dst = dst.GetNode()
	}
	n.l.addEvent(le)
}

// AddEvent adds event whose action will be called on the next loop iteration.
func (n *Node) AddEvent(e event.Actor, dst EventHandler) {
	n.AddEventp(e, dst, elog.PointerToFirstArg(&n))
}

func (n *Node) AddTimedEventp(e event.Actor, dst EventHandler, dt float64, p elog.PointerToFirstArg) {
	le := n.l.getLoopEvent(e, p)
	le.dst = dst.GetNode()
	n.l.addTimedEvent(le, dt)
}
func (n *Node) AddTimedEvent(e event.Actor, dst EventHandler, dt float64) {
	n.AddTimedEventp(e, dst, dt, elog.PointerToFirstArg(&n))
}

func (e *loopEvent) EventAction() {
	if e.dst != nil {
		e.dst.eventWg.Add(1)
		e.dst.rxEvents <- e
	} else {
		e.do()
	}
}

func (e *loopEvent) do() {
	c := e.caller
	if elog.Enabled() {
		if e.dst != nil {
			x := event_action_elog{
				kind:     event_action_elog_start,
				name:     e.dst.elogNodeName,
				sequence: uint32(e.dst.sequence),
			}
			elog.Add(&x)
		}
		c.SetTimeNow()
		if a, ok := e.actor.(elog.Data); ok {
			elog.AddDatac(a, c)
		} else {
			elog.Fc("%s", c, e.actor.String())
		}
	}
	e.actor.EventAction()
	if elog.Enabled() {
		if e.dst != nil {
			x := event_action_elog{
				kind:     event_action_elog_done,
				name:     e.dst.elogNodeName,
				sequence: uint32(e.dst.sequence),
			}
			elog.Add(&x)
			e.dst.sequence++
		}
	}
	e.l.putLoopEvent(e)
}

func (e *loopEvent) String() string { return e.actor.String() }

func (l *Loop) doEvent(e *loopEvent) {
	defer func() {
		if err := recover(); err != nil {
			if err != ErrQuit {
				fmt.Printf("%s: %s", err, debug.Stack())
			}
			l.Quit()
		}
	}()
	e.do()
}

func (l *Loop) eventHandler(p EventHandler) {
	c := p.GetNode()
	// Save elog if thread panics.
	defer func() {
		if elog.Enabled() {
			if err := recover(); err != nil {
				elog.Panic(fmt.Errorf("%s: %v", c.name, err))
				panic(err)
			}
		}
	}()
	for {
		e := <-c.rxEvents
		l.doEvent(e)
		c.eventWg.Done()
	}
}

// If too small, events may block when there are timing mismataches between sender and receiver.
const eventHandlerChanDepth = 16 << 10

func (l *Loop) startHandler(n EventHandler) {
	c := n.GetNode()
	c.rxEvents = make(chan *loopEvent, eventHandlerChanDepth)
	go l.eventHandler(n)
}

func (l *eventMain) eventPoller(p EventPoller) {
	// Save elog if thread panics.
	defer func() {
		if elog.Enabled() {
			if err := recover(); err != nil {
				elog.Panic(err)
				panic(err)
			}
		}
	}()
	for {
		p.EventPoll()
	}
}
func (l *eventMain) startPoller(n EventPoller) { go l.eventPoller(n) }

func (l *Loop) doEventNoWait() (quit *quitEvent, didEvent bool) {
	l.now = cpu.TimeNow()
	select {
	case e := <-l.events:
		didEvent = true
		var ok bool
		if quit, ok = e.actor.(*quitEvent); ok {
			return
		}
		e.EventAction()
	default:
	}
	return
}

func (l *Loop) duration(t cpu.Time) time.Duration {
	return time.Duration(float64(int64(t-l.now)) * l.timeDurationPerCycle)
}

func (l *Loop) doEventWait(dt time.Duration) (quit *quitEvent, didEvent bool) {
	select {
	case e := <-l.events:
		didEvent = true
		var ok bool
		if quit, ok = e.actor.(*quitEvent); ok {
			// Log quit event.
			elog.S("loop quit " + e.String())
		} else {
			e.EventAction()
		}
	case <-time.After(dt):
		elog.S("loop event timeout")
	}
	return
}

func (l *Loop) doEvents() (quitLoop bool) {
	var (
		quit     *quitEvent
		didWait  bool
		didEvent bool
	)
	if _, didWait = l.activePollerState.setEventWait(); didWait {
		l.now = cpu.TimeNow()
		dt := time.Duration(1<<63 - 1)
		if t, ok := l.eventPool.NextTime(); ok {
			dt = l.duration(t)
		}
		if didWait = dt > 0; didWait {
			quit, didEvent = l.doEventWait(dt)
			l.activePollerState.clearEventWait()
		}
	}
	if !didWait {
		quit, didEvent = l.doEventNoWait()
	}

	// Handle expired timed events.
	if l.eventPool.Elts() != 0 {
		l.now = cpu.TimeNow()
		l.eventPoolLock.Lock()
		l.eventPool.AdvanceAdd(l.now, &l.eventVec)
		l.eventPoolLock.Unlock()
		if didEvent = len(l.eventVec) > 0; didEvent {
			for i := range l.eventVec {
				l.eventVec[i].EventAction()
			}
			elog.F2u("timed events %d expired, %d left",
				uint64(len(l.eventVec)), uint64(l.eventPool.Elts()))
			l.eventVec = l.eventVec[:0]
		}
	}

	// Wait for all event handlers to become inactive.
	if didEvent {
		l.eventMain.Wait()
	}

	quitLoop = quit != nil && quit.Type == quitEventExit
	return
}

func (l *eventMain) Wait() {
	for _, h := range l.handlers {
		c := h.GetNode()
		c.eventWg.Wait()
	}
}

func (l *eventMain) Init(p *Loop) {
	l.l = p
	l.events = make(chan *loopEvent, eventHandlerChanDepth)

	for _, n := range l.pollers {
		l.startPoller(n)
	}
	for _, n := range l.handlers {
		p.startHandler(n)
	}
}

func (l *eventMain) RegisterEventPoller(p EventPoller) {
	l.pollers = append(l.pollers, p)
}

type quitEvent struct{ Type quitEventType }
type quitEventType uint8

const (
	quitEventExit quitEventType = iota
	quitEventInterrupt
)

var quitEventTypeStrings = [...]string{
	quitEventExit:      "quit",
	quitEventInterrupt: "interrupt",
}

var (
	ErrQuit      = &quitEvent{Type: quitEventExit}
	ErrInterrupt = &quitEvent{Type: quitEventInterrupt}
)

func (e *quitEvent) String() string { return quitEventTypeStrings[e.Type] }
func (e *quitEvent) Error() string  { return e.String() }
func (e *quitEvent) EventAction()   {}
func (l *Loop) Quit() {
	e := l.getLoopEvent(ErrQuit, elog.PointerToFirstArg(&l))
	l.addEvent(e)
}

// Add an event to wakeup event sleep.
func (l *Loop) Interrupt() {
	e := l.getLoopEvent(ErrInterrupt, elog.PointerToFirstArg(&l))
	l.addEvent(e)
}

const (
	event_action_elog_start = iota
	event_action_elog_done
)

type event_action_elog_kind uint32

func (k event_action_elog_kind) String() string {
	switch k {
	case event_action_elog_start:
		return "start"
	case event_action_elog_done:
		return "done"
	default:
		return fmt.Sprintf("unknown %d", int(k))
	}
}

type event_action_elog struct {
	kind     event_action_elog_kind
	name     elog.StringRef
	sequence uint32
}

func (e *event_action_elog) Elog(l *elog.Log) {
	l.Logf("loop %s%d %s", e.name, e.sequence, e.kind)
}
