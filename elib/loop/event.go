// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loop

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/cpu"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/event"

	"fmt"
	"runtime/debug"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type EventPoller interface {
	EventPoll()
}

type eventMain struct {
	l                 *Loop
	eventPollers      []EventPoller
	eventHandlers     []Noder
	eventHandlerNodes []*Node

	activeNodes   []*Node
	inactiveNodes []*Node

	nodeEventPool sync.Pool
	events        chan *nodeEvent

	// Timed events.
	timer              *time.Timer
	timerCpuTime       cpu.Time
	timerDuration      time.Duration
	timedEventPoolLock sync.Mutex
	timedEventPool     event.Pool
	timedEventVec      event.ActorVec
}

type eventNode struct {
	// Index in active vector if active else ^uint(0).
	activeIndex uint

	activeCount uint32

	// Handler sequence to identify events in event log.
	sequence       uint32
	queue_sequence uint32

	rxEvents chan *nodeEvent

	ft fromToNode

	currentEvent Event
	s            eventNodeState
	eventStats   nodeStats
	activateEvent
}

func (l *eventMain) getLoopEvent(a event.Actor, dst Noder, p elog.PointerToFirstArg) (e *nodeEvent) {
	e = l.nodeEventPool.Get().(*nodeEvent)
	e.d = nil
	e.actor = a
	e.l = l.l
	e.caller = elog.GetCaller(p)
	if dst != nil {
		e.d = dst.GetNode()
		e.d.maybeStartEventHandler()
	}
	return
}
func (l *eventMain) putLoopEvent(x *nodeEvent) { l.nodeEventPool.Put(x) }

type nodeEvent struct {
	l      *Loop
	d      *Node
	actor  event.Actor
	time   cpu.Time
	caller elog.Caller
}

func (e *nodeEvent) EventTime() cpu.Time { return e.time }

func (l *Loop) signalEvent(le *nodeEvent) {
	select {
	case l.events <- le:
	default:
		l.signalEventAfter(le, 0)
	}
}

// SignalEvent adds event whose action will be called on the next loop iteration.
func (n *Node) SignalEventp(a event.Actor, dst Noder, p elog.PointerToFirstArg) {
	e := n.l.getLoopEvent(a, dst, p)
	n.l.signalEvent(e)
}

// SignalEvent adds event whose action will be called on the next loop iteration.
func (n *Node) SignalEvent(e event.Actor, dst Noder) {
	n.SignalEventp(e, dst, elog.PointerToFirstArg(&n))
}

func (l *Loop) signalEventAfter(le *nodeEvent, secs float64) {
	// For first signal use current time; for re-signals use time after last signal.
	if le.time == 0 {
		le.time = cpu.TimeNow()
	}
	le.time += cpu.Time(secs * l.cyclesPerSec)
	l.timedEventPoolLock.Lock()
	defer l.timedEventPoolLock.Unlock()
	l.timedEventPool.Add(le)
}

func (n *Node) SignalEventAfterp(a event.Actor, dst Noder, dt float64, p elog.PointerToFirstArg) {
	e := n.l.getLoopEvent(a, dst, p)
	n.l.signalEventAfter(e, dt)
}
func (n *Node) SignalEventAfter(e event.Actor, dst Noder, secs float64) {
	n.SignalEventAfterp(e, dst, secs, elog.PointerToFirstArg(&n))
}

func (e *nodeEvent) logActor() {
	c := e.caller
	c.SetTimeNow()
	if a, ok := e.actor.(elog.Data); ok {
		elog.AddDatac(a, c)
	} else {
		elog.Fc("%s", c, e.actor.String())
	}
}

func (e *nodeEvent) do() {
	d, n := e.d, &e.d.e
	if elog.Enabled() {
		n.log(d, event_elog_action)
		e.logActor()
	}

	if a, ok := e.actor.(EventActor); ok {
		x := a.getLoopEvent()
		x.e = e
	}

	t0 := cpu.TimeNow()
	e.actor.EventAction()
	e.d.e.eventStats.update(1, t0)
	n.log(d, event_elog_action_done)
	n.sequence++ // done => use next sequence
	e.l.putLoopEvent(e)
}

func (e *nodeEvent) String() string { return e.actor.String() }

func (d *Node) eventDone() {
	n := &d.e
	n.s.setDone(d)
	n.currentEvent.e = nil
	n.activeCount--
	n.log(d, event_elog_node_signal_done)
	n.ft.signalLoop(true)
}

func (l *Loop) eventHandler(r Noder) {
	d := r.GetNode()
	// Save elog if thread panics.
	defer func() {
		if err := recover(); err != nil {
			if err == ErrQuit {
				l.Quit()
				return
			}
			err = fmt.Errorf("%v: %v", d.name, err)
			elog.Panic(err)
			l.Panic(err, debug.Stack())
			d.eventDone()
		}
	}()
	n := &d.e
	for {
		n.log(d, event_elog_node_wait)
		n.ft.waitLoop()
		n.log(d, event_elog_node_wake)
		e := <-n.rxEvents
		if poller_panics && e.d != d {
			panic(fmt.Errorf("expected node %s got %s: %p %s", d.name, e.d.name, e, e.actor.String()))
		}
		n.currentEvent.e = e
		e.do()
		d.eventDone()
	}
}

// Types capable will include declare loop.Event and thereby inherit Suspend/Resume.
type Event struct {
	e *nodeEvent
}

type EventActor interface {
	getLoopEvent() *Event
}

func (e *Event) getLoopEvent() *Event { return e }
func (n *Node) CurrentEvent() (e *Event) {
	x := &n.e.currentEvent
	if x.e != nil {
		e = x
	}
	return
}

func (x *Event) Suspend() {
	d := x.e.d
	n := &d.e
	if !n.isActive() {
		panic("suspending inactive node")
	}
	if was := n.s.setSuspend(d, true); was {
		n.logsi(d, event_elog_suspend, n.sequence, "ignore duplicate suspend")
		return
	}
	n.log(d, event_elog_suspend)
	n.eventStats.current.suspends++
	t0 := cpu.TimeNow()
	n.ft.signalLoop(false)
	n.ft.waitLoop()
	// Don't charge node for time suspended.
	dt := cpu.TimeNow() - t0
	n.eventStats.current.clocks -= uint64(dt)
	n.log(d, event_elog_resumed)
}

func (e *nodeEvent) isResume() bool { return e.actor == nil }
func (e *nodeEvent) setResume()     { e.actor = nil }

func (x *Event) Resume() (ok bool) {
	e := x.e
	d := e.d
	n := &d.e

	// Don't do it twice.
	if ok, _, _ = n.s.setResume(d); !ok {
		n.logsi(d, event_elog_queue_resume, n.sequence, "ignore duplicate resume")
		return
	}
	n.log(d, event_elog_queue_resume)
	e.setResume()
	d.l.events <- e
	return
}

func (x *nodeEvent) resume() {
	d, n := x.d, &x.d.e
	n.log(d, event_elog_resume_wake)
	n.s.clearResume(d)
}

// If too small, events may block when there are timing mismataches between sender and receiver.
const eventHandlerChanDepth = 1 << 10

func (n *Node) hasEventHandler() bool { return n.e.rxEvents != nil }
func (d *Node) maybeStartEventHandler() {
	n := &d.e
	if n.rxEvents != nil {
		return
	}
	l := d.l
	l.eventHandlers = append(l.eventHandlers, d.noder)
	l.eventHandlerNodes = append(l.eventHandlerNodes, d)
	n.rxEvents = make(chan *nodeEvent, eventHandlerChanDepth)
	n.activeIndex = ^uint(0)
	n.ft.init()
	elog.F("loop starting event handler %v", d.elogNodeName)
	go l.eventHandler(d.noder)
}

func (l *Loop) eventPoller(p EventPoller) {
	// Save elog if thread panics.
	defer func() {
		if elog.Enabled() {
			if err := recover(); err != nil {
				elog.Panic(err)
				err = fmt.Errorf("event-poller: %v", err)
				elog.Panic(err)
				l.Panic(err, debug.Stack())
			}
		}
	}()
	for {
		p.EventPoll()
	}
}
func (l *Loop) startEventPoller(n EventPoller)         { go l.eventPoller(n) }
func (l *eventMain) RegisterEventPoller(p EventPoller) { l.eventPollers = append(l.eventPollers, p) }

func (e *nodeEvent) EventAction() {
	d := e.d
	if d == nil { // this can happen with timed activateEvent.
		e.actor.EventAction()
		return
	}
	n := &d.e

	// Set signal time for timed events.
	if e.time != 0 {
		e.time = d.l.now
	}

	if elog.Enabled() {
		n.logsi(d, event_elog_queue, n.queue_sequence, e.actor.String())
		n.queue_sequence++
	}
	n.activeCount++
	if n.activeCount == 1 {
		d.l.eventMain.addActive(d)
	}
	n.rxEvents <- e
}

func (m *eventMain) doNodeEvent(e *nodeEvent) (quit *quitEvent) {
	var ok bool
	if quit, ok = e.actor.(*quitEvent); ok {
		return
	}
	if e.isResume() {
		m.addActive(e.d)
		e.resume()
	} else {
		e.EventAction()
	}
	return
}

func (l *Loop) doEventNoWait() (quit *quitEvent) {
	select {
	default: // nothing to do
	case e := <-l.events:
		quit = l.doNodeEvent(e)
	}
	return
}

func (l *Loop) doEventWait() (quit *quitEvent, timeout bool) {
	m := &l.eventMain
	m.event_timer_elog(event_timer_elog_waiting, m.timerDuration)
	select {
	case e := <-l.events:
		quit = l.doNodeEvent(e)
	case <-m.timer.C:
		// Log difference between time now and timer cpu time.
		m.event_timer_elog(event_timer_elog_timeout, l.duration(m.timerCpuTime))
		m.timer.Reset(maxDuration)
		timeout = true
	}
	return
}

func (l *Loop) duration(t cpu.Time) time.Duration {
	l.now = cpu.TimeNow()
	return time.Duration(float64(int64(t-l.now)) * l.timeDurationPerCycle)
}

func (l *Loop) doEvents() (quitLoop bool) {
	m := &l.eventMain
	var (
		quit          *quitEvent
		didWait       bool
		waitTimeout   bool
		nextTimeValid bool
		nextTime      cpu.Time
	)

	// Try waiting if we have no active nodes.
	if len(m.activeNodes) == 0 {
		// Try to change active poller state to event wait.
		// This can and does return false if an active poller comes along racing with our call.
		if _, didWait = l.activePollerState.setEventWait(); didWait {
			// Find next event's time (!ok means there is no available event).
			nextTime, nextTimeValid = l.timedEventPool.NextTime()

			// Compute duration until next event.
			var dt time.Duration
			if nextTimeValid {
				dt = l.duration(nextTime)
			} else {
				nextTime = maxCpuTime
				dt = maxDuration
			}

			// Reset timer if wakeup time changes.
			if nextTime != m.timerCpuTime {
				if !m.timer.Stop() {
					<-m.timer.C
				}
				m.timer.Reset(dt)
				m.timerCpuTime = nextTime
				m.timerDuration = dt
				m.event_timer_elog(event_timer_elog_reset, dt)
			}
			quit, waitTimeout = l.doEventWait()
			l.activePollerState.clearEventWait()
		}
	}
	if !didWait {
		quit = l.doEventNoWait()
	}

	// Handle expired timed events.
	tp := &l.timedEventPool
	if waitTimeout {
		l.timedEventPoolLock.Lock()
		ev := l.timedEventVec
		tp.AdvanceAdd(nextTime, &ev)
		l.timedEventPoolLock.Unlock()
		if poller_panics && waitTimeout && len(ev) == 0 {
			panic("wait timeout but not events expired")
		}
		if len(ev) > 0 {
			if elog.Enabled() {
				elog.F2u("loop event timer %d expired, %d queued",
					uint64(len(ev)), uint64(tp.Elts()))
			}
			for i := range ev {
				ev[i].EventAction()
			}
			// Save away for next use.
			l.timedEventVec = ev[:0]
		}
	}

	// Signal all active nodes to start.
	for _, d := range m.activeNodes {
		n := &d.e
		n.log(d, event_elog_start)
		n.ft.signalNode()
	}

	// Wait for all event active nodes to finish.
	for _, d := range m.activeNodes {
		n := &d.e
		q := n.sequence
		n.log(d, event_elog_wait)
		nodeEventDone := n.ft.waitNode()
		// Inactivate nodes which have no more queued events or are suspended.
		if !nodeEventDone || n.activeCount == 0 {
			m.inactiveNodes = append(m.inactiveNodes, d)
		}
		n.logi(d, event_elog_wait_done, q)
	}

	if len(m.inactiveNodes) > 0 {
		for _, d := range m.inactiveNodes {
			m.delActive(d)
		}
		m.inactiveNodes = m.inactiveNodes[:0]
	}

	quitLoop = (quit != nil && quit.Type == quitEventExit) || l.isPanic()
	return
}

func (m *eventMain) addActive(d *Node) {
	n := &d.e
	if n.isActive() {
		n.logsi(d, event_elog_add_active, n.sequence, "ignore duplicate")
		return
	}
	n.activeIndex = uint(len(m.activeNodes))
	m.activeNodes = append(m.activeNodes, d)
	n.logi(d, event_elog_add_active, uint32(len(m.activeNodes)))
}

func (n *eventNode) isActive() bool { return n.activeIndex != ^uint(0) }
func (m *eventMain) delActive(d *Node) {
	n := &d.e
	ai := n.activeIndex
	l := uint(len(m.activeNodes))
	if l > 0 && ai < l-1 {
		m.activeNodes[ai] = m.activeNodes[l-1]
	}
	m.activeNodes = m.activeNodes[:l-1]
	n.activeIndex = ^uint(0)
	n.logi(d, event_elog_del_active, uint32(len(m.activeNodes)))
}

type eventNodeState uint32

func (t eventNodeState) String() (s string) {
	if t&1 != 0 {
		s += "suspended"
	}
	if t&2 != 0 {
		s += "resumed"
	}
	if s == "" {
		s = "active"
	}
	return
}

const (
	// Logged by main loop.
	event_node_state_elog_suspend = iota
	event_node_state_elog_set_resume
	event_node_state_elog_clear_resume
)

type event_node_state_elog_kind uint32

func (k event_node_state_elog_kind) String() string {
	t := [...]string{
		event_node_state_elog_suspend:      "suspend",
		event_node_state_elog_set_resume:   "set-resume",
		event_node_state_elog_clear_resume: "clear-resume",
	}
	return elib.StringerHex(t[:], int(k))
}

type event_node_state_elog struct {
	kind     event_node_state_elog_kind
	name     elog.StringRef
	old, new eventNodeState
}

func (e *event_node_state_elog) Elog(l *elog.Log) {
	l.Logf("event node state %v %v %v -> %v", e.kind, e.name, e.old, e.new)
}

func (s *eventNodeState) compare_and_swap(old, new eventNodeState) (swapped bool) {
	return atomic.CompareAndSwapUint32((*uint32)(s), uint32(old), uint32(new))
}
func (s *eventNodeState) get() (x eventNodeState, isSuspended, isResumed bool) {
	x = eventNodeState(atomic.LoadUint32((*uint32)(s)))
	isSuspended = x&1 != 0
	isResumed = x&2 != 0
	return
}
func makeEventNodeState(isSuspended, isResumed bool) (s eventNodeState) {
	if isSuspended {
		s |= 1
	}
	if isResumed {
		s |= 2
	}
	return
}
func (s *eventNodeState) setDone(d *Node) { s.setSuspend(d, false) }
func (s *eventNodeState) setSuspend(d *Node, is bool) (was bool) {
	for {
		var old eventNodeState
		old, was, _ = s.get()
		if is == was {
			return
		}
		new := makeEventNodeState(is, false)
		if s.compare_and_swap(old, new) {
			if elog.Enabled() {
				elog.Add(&event_node_state_elog{
					kind: event_node_state_elog_suspend,
					name: d.elogNodeName,
					old:  old,
					new:  new,
				})
			}
			return
		}
	}
}
func (s *eventNodeState) isResumed() (ok bool)   { _, _, ok = s.get(); return }
func (s *eventNodeState) isSuspended() (ok bool) { _, ok, _ = s.get(); return }
func (s *eventNodeState) setResume(d *Node) (ok, wasSuspended, wasResumed bool) {
	var old eventNodeState
	if old, wasSuspended, wasResumed = s.get(); wasSuspended && !wasResumed {
		new := makeEventNodeState(false, true)
		ok = s.compare_and_swap(old, new)
		if ok {
			elog.Add(&event_node_state_elog{
				kind: event_node_state_elog_set_resume,
				name: d.elogNodeName,
				old:  old,
				new:  new,
			})
		}
	}
	return
}
func (s *eventNodeState) clearResume(d *Node) bool {
	for {
		old, wasSuspended, wasResumed := s.get()
		if !wasResumed {
			return wasResumed
		}
		new := makeEventNodeState(wasSuspended, false)
		if s.compare_and_swap(old, new) {
			elog.Add(&event_node_state_elog{
				kind: event_node_state_elog_clear_resume,
				name: d.elogNodeName,
				old:  old,
				new:  new,
			})
			return wasResumed
		}
	}
}

const (
	maxDuration = 1<<63 - 1
	// Cpu time indicating that timer is armed with maxDuration.
	maxCpuTime = ^cpu.Time(0)
)

func (m *eventMain) eventInit(l *Loop) {
	m.l = l
	m.events = make(chan *nodeEvent, eventHandlerChanDepth)
	m.timerCpuTime = maxCpuTime
	m.timerDuration = maxDuration
	m.timer = time.NewTimer(maxDuration)
	m.nodeEventPool.New = func() interface{} { return &nodeEvent{} }
	m.event_timer_elog(event_timer_elog_reset, maxDuration)
	for _, n := range l.eventPollers {
		l.startEventPoller(n)
	}
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
	e := l.getLoopEvent(ErrQuit, nil, elog.PointerToFirstArg(&l))
	l.signalEvent(e)
}

// Add an event to wakeup event sleep.
func (l *Loop) Interrupt() {
	e := l.getLoopEvent(ErrInterrupt, nil, elog.PointerToFirstArg(&l))
	l.signalEvent(e)
}

func (l *Loop) showRuntimeEvents(w cli.Writer) (err error) {
	type event struct {
		Name     string  `format:"%-30s"`
		Events   uint64  `format:"%16d"`
		Suspends uint64  `format:"%16d"`
		Clocks   float64 `format:"%16.2f"`
	}

	es := []event{}
	var inputSummary stats
	for _, n := range l.nodes {
		if !n.hasEventHandler() {
			continue
		}
		var s stats
		s.add(&n.e.eventStats)
		inputSummary.add(&n.e.eventStats)
		es = append(es, event{
			Name:     n.name,
			Events:   s.vectors,
			Suspends: s.suspends,
			Clocks:   s.clocksPerVector(),
		})
	}

	// Summary
	if s := inputSummary; s.calls > 0 {
		dt := time.Since(l.timeLastRuntimeClear).Seconds()
		eventsPerSec := float64(s.vectors) / dt
		clocksPerEvent := float64(s.clocks) / float64(s.vectors)
		fmt.Fprintf(w, "Events: %d, Events/sec: %.2e, Clocks/event: %.2f\n",
			s.vectors, eventsPerSec, clocksPerEvent)
	}

	sort.Slice(es, func(i, j int) bool { return es[i].Name < es[j].Name })
	elib.TabulateWrite(w, es)
	return
}

const (
	// Logged by main loop.
	event_elog_queue = iota
	event_elog_start
	event_elog_wait
	event_elog_wait_done
	event_elog_add_active
	event_elog_del_active
	event_elog_suspend_wake
	event_elog_resume_wake
	// Logged by node.
	event_elog_node_wake
	event_elog_node_wait
	event_elog_node_signal_done
	event_elog_action
	event_elog_action_done
	event_elog_suspend
	event_elog_resumed
	event_elog_queue_resume
)

type event_elog_kind uint32

func (k event_elog_kind) String() string {
	t := [...]string{
		event_elog_queue:            "queue",
		event_elog_start:            "start",
		event_elog_wait:             "wait",
		event_elog_wait_done:        "wait-done",
		event_elog_add_active:       "add-active",
		event_elog_del_active:       "del-active",
		event_elog_resume_wake:      "resume-wake",
		event_elog_suspend_wake:     "suspend-wake",
		event_elog_node_wait:        "wait",
		event_elog_node_wake:        "wake",
		event_elog_node_signal_done: "signal-done",
		event_elog_action:           "action",
		event_elog_action_done:      "action-done",
		event_elog_suspend:          "suspend",
		event_elog_resumed:          "resumed",
		event_elog_queue_resume:     "queue-resume",
	}
	return elib.StringerHex(t[:], int(k))
}

func (n *eventNode) logsi(d *Node, kind event_elog_kind, i uint32, s string) {
	if elog.Enabled() {
		e := event_elog{
			name: d.elogNodeName,
			kind: kind,
			i:    i,
		}
		copy(e.s[:], []byte(s))
		elog.Add(&e)
	}
}
func (n *eventNode) logi(d *Node, kind event_elog_kind, i uint32) { n.logsi(d, kind, i, "") }
func (n *eventNode) log(d *Node, kind event_elog_kind)            { n.logi(d, kind, n.sequence) }

type event_elog struct {
	kind event_elog_kind
	name elog.StringRef
	i    uint32
	s    [elog.EventDataBytes - 3*4]byte
}

func (e *event_elog) Elog(l *elog.Log) {
	s := elog.String(e.s[:])
	if s != "" {
		s = ": " + s
	}
	switch e.kind {
	case event_elog_node_wake, event_elog_node_wait,
		event_elog_action, event_elog_action_done,
		event_elog_suspend, event_elog_resumed, event_elog_queue_resume:
		// Events generated by node.
		l.Logf("loop event node %v %s %d%s", e.name, e.kind, e.i, s)
	default:
		switch {
		case e.i == ^uint32(0):
			l.Logf("loop event %v %v%s", e.kind, e.name, s)
		case e.kind == event_elog_add_active || e.kind == event_elog_del_active:
			l.Logf("loop event %v %v %d%s", e.kind, e.name, e.i, s)
		default:
			l.Logf("loop event %v %v %d%s", e.kind, e.name, e.i, s)
		}
	}
}

const (
	event_timer_elog_waiting = iota
	event_timer_elog_reset
	event_timer_elog_timeout
)

type event_timer_elog_kind uint32

func (k event_timer_elog_kind) String() string {
	t := [...]string{
		event_timer_elog_waiting: "waiting",
		event_timer_elog_reset:   "reset",
		event_timer_elog_timeout: "timeout",
	}
	return elib.StringerHex(t[:], int(k))
}

type event_timer_elog struct {
	kind event_timer_elog_kind
	dt   time.Duration
}

func (e *event_timer_elog) Elog(l *elog.Log) {
	switch e.kind {
	case event_timer_elog_timeout:
		l.Logf("loop event timer %v error %+.2e", e.kind, e.dt.Seconds())
	default:
		if e.dt == maxDuration {
			l.Logf("loop event timer %v forever", e.kind)
		} else {
			l.Logf("loop event timer %v %.2e sec", e.kind, e.dt.Seconds())
		}
	}
}

func (m *eventMain) event_timer_elog(kind event_timer_elog_kind, dt time.Duration) {
	if elog.Enabled() {
		e := event_timer_elog{kind: kind, dt: dt}
		elog.Add(&e)
	}
}
