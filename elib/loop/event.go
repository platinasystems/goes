// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loop

import (
	"github.com/platinasystems/go/elib/cpu"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/event"

	"fmt"
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

type eventLoop struct {
	l        *Loop
	pollers  []EventPoller
	handlers []EventHandler

	loopEvents    sync.Pool
	events        chan *loopEvent
	eventPoolLock sync.Mutex
	eventPool     event.Pool
	eventVec      event.ActorVec
}

func (l *eventLoop) getLoopEvent(a event.Actor) (x *loopEvent) {
	if y := l.loopEvents.Get(); y != nil {
		x = y.(*loopEvent)
		*x = loopEvent{actor: a}
	} else {
		x = &loopEvent{actor: a}
	}
	x.l = l.l
	return
}
func (l *eventLoop) putLoopEvent(x *loopEvent) { l.loopEvents.Put(x) }

type eventNode struct {
	activateEvent
	rxEvents chan *loopEvent
	eventVec event.ActorVec
}

type loopEvent struct {
	l     *Loop
	actor event.Actor
	dst   *Node
	time  cpu.Time
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
func (n *Node) AddEvent(e event.Actor, dst EventHandler) {
	le := n.loop.getLoopEvent(e)
	if dst != nil {
		le.dst = dst.GetNode()
	}
	// Never block when polling.
	blocking := !n.is_polling()
	n.loop.addEvent(le, blocking)
}

func (n *Node) AddTimedEvent(e event.Actor, dst EventHandler, dt float64) {
	le := n.loop.getLoopEvent(e)
	le.dst = dst.GetNode()
	n.loop.addTimedEvent(le, dt)
}

func (e *loopEvent) EventAction() {
	if e.dst != nil {
		e.dst.rxEvents <- e
		e.dst.set_active(true)
	} else {
		e.do()
	}
}

func (e *loopEvent) do() {
	if elog.Enabled() {
		le := eventElogEvent{}
		copy(le.s[:], e.String())
		le.Log()
	}
	e.actor.EventAction()
	e.l.putLoopEvent(e)
}

func (e *loopEvent) String() string { return e.actor.String() }

func (l *Loop) doEvent(e *loopEvent) {
	defer func() {
		if err := recover(); err == ErrQuit {
			l.Quit()
		} else if err != nil {
			fmt.Println(err)
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
	c.toLoop = make(chan struct{}, 1)
	c.fromLoop = make(chan struct{}, 1)
	c.rxEvents = make(chan *loopEvent, eventHandlerChanDepth)
	go l.eventHandler(n)
}

func (l *eventLoop) eventPoller(p EventPoller) {
	for {
		p.EventPoll()
	}
}
func (l *eventLoop) startPoller(n EventPoller) { go l.eventPoller(n) }

func (l *Loop) doEventNoWait() (quit *quitEvent) {
	l.now = cpu.TimeNow()
	select {
	case e := <-l.events:
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

func (l *Loop) doEventWait() (quit *quitEvent) {
	l.now = cpu.TimeNow()
	dt := time.Duration(1<<63 - 1)
	if t, ok := l.eventPool.NextTime(); ok {
		if dt = l.duration(t); dt <= 0 {
			return
		}
	}
	if elog.Enabled() {
		elog.GenEvent("waiting for event")
	}
	l.waitingForEvent = true
	select {
	case e := <-l.events:
		var ok bool
		if quit, ok = e.actor.(*quitEvent); ok {
			// Log quit event.
			if elog.Enabled() {
				elog.GenEvent("wakeup " + e.String())
			}
		} else {
			e.EventAction()
		}
	case <-time.After(dt):
		if elog.Enabled() {
			elog.GenEvent("wakeup timeout")
		}
	}
	l.waitingForEvent = false
	return
}

func (l *Loop) doEvents() (quitLoop bool) {
	// Handle discrete events.
	var quit *quitEvent
	if l.nActivePollers > 0 {
		quit = l.doEventNoWait()
	} else {
		quit = l.doEventWait()
	}
	if quit != nil {
		quitLoop = quit.Type == quitEventExit
		return
	}

	// Handle expired timed events.
	l.eventPoolLock.Lock()
	l.eventPool.AdvanceAdd(l.now, &l.eventVec)
	l.eventPoolLock.Unlock()

	for i := range l.eventVec {
		l.eventVec[i].EventAction()
	}
	l.eventVec = l.eventVec[:0]

	// Wait for all event handlers to become inactive.
	l.eventLoop.Wait()

	return
}

func (l *eventLoop) Wait() {
	for _, h := range l.handlers {
		c := h.GetNode()
		if c.is_active() {
			<-c.toLoop
			c.set_active(false)
		}
	}
}

func (p *Loop) eventInit() {
	l := &p.eventLoop
	l.l = p
	l.events = make(chan *loopEvent, eventHandlerChanDepth)

	for _, n := range l.pollers {
		l.startPoller(n)
	}
	for _, n := range l.handlers {
		p.startHandler(n)
	}
}

func (l *eventLoop) RegisterEventPoller(p EventPoller) {
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
func (l *Loop) Quit()               { l.addEvent(l.getLoopEvent(ErrQuit), true) }
func (l *Loop) Interrupt() {
	// Add an event to wakeup event sleep.
	if l.waitingForEvent {
		l.addEvent(l.getLoopEvent(ErrInterrupt), false)
	}
}
