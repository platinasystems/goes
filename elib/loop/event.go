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
	rxEvents      chan *loopEvent
	nActiveEvents uint32
	eventVec      event.ActorVec
}

type loopEvent struct {
	l      *Loop
	actor  event.Actor
	caller elog.Caller
	dst    *Node
	time   cpu.Time
}

func (e *loopEvent) EventTime() cpu.Time { return e.time }

func (l *Loop) addEvent(le *loopEvent, blocking bool) {
	if blocking {
		l.events <- le
	} else {
		select {
		case l.events <- le:
		default:
			l.addTimedEvent(le, 0)
		}
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
	le := n.loop.getLoopEvent(e, p)
	if dst != nil {
		le.dst = dst.GetNode()
	}
	// Never block when polling.
	blocking := !n.is_polling()
	n.loop.addEvent(le, blocking)
}

// AddEvent adds event whose action will be called on the next loop iteration.
func (n *Node) AddEvent(e event.Actor, dst EventHandler) {
	n.AddEventp(e, dst, elog.PointerToFirstArg(&n))
}

func (n *Node) AddTimedEventp(e event.Actor, dst EventHandler, dt float64, p elog.PointerToFirstArg) {
	le := n.loop.getLoopEvent(e, p)
	le.dst = dst.GetNode()
	n.loop.addTimedEvent(le, dt)
}
func (n *Node) AddTimedEvent(e event.Actor, dst EventHandler, dt float64) {
	n.AddTimedEventp(e, dst, dt, elog.PointerToFirstArg(&n))
}

func (e *loopEvent) EventAction() {
	if e.dst != nil {
		e.dst.nActiveEvents++
		e.dst.rxEvents <- e
	} else {
		e.do()
	}
}

func (e *loopEvent) do() {
	if elog.Enabled() {
		if a, ok := e.actor.(elog.Data); ok {
			elog.AddDatac(a, e.caller)
		} else {
			elog.Sc(e.actor.String(), e.caller)
		}
	}
	e.actor.EventAction()
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
	for {
		e := <-c.rxEvents
		l.doEvent(e)
		c.toLoop <- struct{}{}
	}
}

// If too small, events may block when there are timing mismataches between sender and receiver.
const eventHandlerChanDepth = 16 << 10

func (l *Loop) startHandler(n EventHandler) {
	c := n.GetNode()
	c.toLoop = make(chan struct{}, 64)
	c.fromLoop = make(chan struct{}, 1)
	c.rxEvents = make(chan *loopEvent, eventHandlerChanDepth)
	go l.eventHandler(n)
}

func (l *eventMain) eventPoller(p EventPoller) {
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
	elog.S("loop event wait")
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
		for c.nActiveEvents > 0 {
			c.nActiveEvents--
			<-c.toLoop
		}
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
	l.addEvent(e, true)
}

// Add an event to wakeup event sleep.
func (l *Loop) Interrupt() {
	e := l.getLoopEvent(ErrInterrupt, elog.PointerToFirstArg(&l))
	l.addEvent(e, false)
}
