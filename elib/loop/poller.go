// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loop

import (
	"github.com/platinasystems/go/elib/cpu"
	"github.com/platinasystems/go/elib/elog"

	"fmt"
	"sync"
	"sync/atomic"
)

type nodeState struct {
	active       int32
	suspend      int32
	is_active    bool
	is_suspended bool
	is_pending   bool
}

func (t *nodeState) String() (s string) {
	if t.is_active {
		s += "Active"
	}
	if t.is_suspended {
		s += "Suspend"
	}
	if s != "" {
		s += fmt.Sprintf(" %d/%d", t.active, t.suspend)
	}
	return
}

type nodeActivePending struct {
	nodeIndex    uint
	is_active    bool
	is_suspended bool
}

type nodeStateMain struct {
	mu sync.Mutex

	sequence uint

	activePending [2][]nodeActivePending
}

type SuspendLimits struct {
	Suspend, Resume int
}

func (n *Node) addActivity(da, ds int, lim *SuspendLimits) (was_active, did_suspend, did_resume bool) {
	m := &n.l.nodeStateMain
	m.mu.Lock()
	was_active, did_suspend, did_resume = n.addActivityNoLock(int32(da), int32(ds), lim)
	m.mu.Unlock()
	return
}

func (n *Node) addActivityNoLock(da, ds int32, lim *SuspendLimits) (was_active, did_suspend, did_resume bool) {
	m := &n.l.nodeStateMain
	s := &n.s
	active := atomic.AddInt32(&s.active, da)
	suspend := atomic.AddInt32(&s.suspend, ds)
	if active < 0 {
		panic("active < 0")
	}
	if suspend < 0 {
		panic("suspend < 0")
	}
	is_active := active > 0
	if lim == nil {
		n.poller_elog_a(poller_elog_data_activity, da, active)
	} else {
		limit := int32(lim.Suspend)
		if s.is_suspended {
			limit = int32(lim.Resume)
		}
		is_active = is_active && suspend <= limit
		if did_suspend = ds > 0 && active > 0 && suspend > limit; did_suspend {
			// Back-up so suspend count is never above limit.
			atomic.AddInt32(&s.suspend, -ds)
			suspend -= ds
		}
		did_resume = s.is_suspended && is_active
		s.is_suspended = did_suspend && !did_resume
		n.poller_elog_a(poller_elog_suspend_activity, ds, suspend)
	}
	if was_active = s.is_active; was_active != is_active {
		s.is_active = is_active
		n.addPending(m)
		if _, eventWait := n.l.activePollerState.changeActive(is_active); eventWait {
			n.poller_elog(poller_elog_event_wake)
			n.l.Interrupt()
		}
	}
	return
}

func (n *Node) addPending(m *nodeStateMain) {
	s := &n.s
	if !s.is_pending {
		s.is_pending = true
		i := m.sequence & 1
		n.poller_elog(poller_elog_add_pending)
		m.activePending[i] = append(m.activePending[i],
			nodeActivePending{
				nodeIndex:    n.index,
				is_active:    s.is_active,
				is_suspended: s.is_suspended,
			})
	}
}

func (n *Node) AddDataActivity(i int) { n.addActivity(i, 0, nil) }
func (n *Node) Activate(enable bool) (was bool) {
	m := &n.l.nodeStateMain
	s := &n.s
	m.mu.Lock()
	was = s.is_active
	if was != enable {
		da := int32(1)
		if !enable {
			da = -1
		}
		n.addActivityNoLock(da, 0, nil)
	}
	m.mu.Unlock()
	return
}
func (n *Node) IsActive() bool { return n.s.is_active }

func (m *nodeStateMain) advance(nodes []Noder) (pending []nodeActivePending) {
	m.mu.Lock()
	i0 := m.sequence & 1
	i1 := i0 ^ 1

	// Reset pending for next sequence.
	if m.activePending[i1] != nil {
		m.activePending[i1] = m.activePending[i1][:0]
	}
	pending = m.activePending[i0]
	// Clear pending state while we still have lock.
	for _, p := range pending {
		n := nodes[p.nodeIndex].GetNode()
		n.s.is_pending = false
	}
	m.sequence++
	m.mu.Unlock()
	return
}

type activateEvent struct{ n *Node }

func (e *activateEvent) EventAction()   { e.n.Activate(true) }
func (e *activateEvent) String() string { return fmt.Sprintf("activate %s", e.n.name) }

func (n *Node) ActivateAfterTime(dt float64) {
	if was := n.Activate(false); was {
		n.activateEvent.n = n
		le := n.l.getLoopEvent(&n.activateEvent, elog.PointerToFirstArg(&n))
		n.l.addTimedEvent(le, dt)
	}
}

func (l *Loop) AddSuspendActivity(in *In, i int, lim *SuspendLimits) (did_suspend bool, did_resume bool) {
	a := l.activePollerPool.entries[in.activeIndex]
	n := a.pollerNode
	// Node may be nil when inactive and freed but still having outstanding suspend activity.
	if n == nil {
		return
	}
	_, did_suspend, did_resume = n.addActivity(0, i, lim)
	if did_suspend {
		// Signal polling done to main loop.
		n.inputStats.current.suspends++
		n.toLoop <- struct{}{}
		n.poller_elog(poller_elog_suspended)
		// Wait for continue (resume) signal from main loop.
		t0 := cpu.TimeNow()
		<-n.fromLoop
		// Don't charge node for time suspended.
		dt := cpu.TimeNow() - t0
		n.outputStats.current.clocks -= dt
		n.poller_elog(poller_elog_resumed)
	}
	return
}

func (l *Loop) Suspend(in *In, lim *SuspendLimits) { l.AddSuspendActivity(in, 1, lim) }
func (l *Loop) Resume(in *In, lim *SuspendLimits)  { l.AddSuspendActivity(in, -1, lim) }

type activePollerState uint32

func (s *activePollerState) compare_and_swap(old, new activePollerState) (swapped bool) {
	return atomic.CompareAndSwapUint32((*uint32)(s), uint32(old), uint32(new))
}
func (s *activePollerState) get() (x activePollerState, nActive uint, eventWait bool) {
	x = activePollerState(atomic.LoadUint32((*uint32)(s)))
	eventWait = x&1 != 0
	nActive = uint(x >> 1)
	return
}
func makeActivePollerState(nActive uint, eventWait bool) (s activePollerState) {
	s = activePollerState(nActive << 1)
	if eventWait {
		s |= 1
	}
	return
}
func (s *activePollerState) setEventWait() (nActive uint, wait bool) {
	var old activePollerState
	if old, nActive, wait = s.get(); nActive == 0 {
		wantWait := true
		new := makeActivePollerState(nActive, wantWait)
		if !s.compare_and_swap(old, new) {
			return
		}
		wait = wantWait
	}
	return
}
func (s *activePollerState) clearEventWait() {
	old, nActive, wait := s.get()
	for wait {
		new := makeActivePollerState(nActive, false)
		if s.compare_and_swap(old, new) {
			break
		}
		old, nActive, wait = s.get()
	}
}

func (s *activePollerState) changeActive(isActive bool) (uint, bool) {
	for {
		old, n, w := s.get()
		if isActive {
			n += 1
		} else {
			if n == 0 {
				panic("negative active count")
			}
			n -= 1
		}
		new := makeActivePollerState(n, w && n == 0)
		if s.compare_and_swap(old, new) {
			return n, w
		}
	}
}

func (n *Node) getActivePoller(l *Loop) *activePoller {
	return l.activePollerPool.entries[n.activePollerIndex]
}

func (n *Node) allocActivePoller(l *Loop) {
	// Already allocated?
	if !l.activePollerPool.IsFree(n.activePollerIndex) {
		return
	}
	i := l.activePollerPool.GetIndex()
	a := l.activePollerPool.entries[i]
	if a == nil {
		a = &activePoller{}
		l.activePollerPool.entries[i] = a
	}
	a.index = uint16(i)
	n.activePollerIndex = i
	a.pollerNode = n
	n.poller_elog_i(poller_elog_alloc, i)
}

func (n *Node) freeActivePoller(l *Loop) {
	a := n.getActivePoller(l)
	a.flushNodeStats(l)
	a.pollerNode = nil
	i := n.activePollerIndex
	l.activePollerPool.PutIndex(i)
	n.activePollerIndex = ^uint(0)
	n.poller_elog_i(poller_elog_free, i)
}

func (a *activePoller) flushNodeStats(l *Loop) {
	for i := range a.activeNodes {
		an := &a.activeNodes[i]
		n := l.DataNodes[an.index].GetNode()

		n.inputStats.current.add_raw(&an.inputStats)
		an.inputStats.zero()

		n.outputStats.current.add_raw(&an.outputStats)
		an.outputStats.zero()
	}
}

func (l *Loop) dataPoll(p inLooper) {
	defer func() {
		if err := recover(); err != nil {
			if elog.Enabled() {
				elog.Panic(err)
			}
			panic(err)
		}
	}()
	c := p.GetNode()
	for {
		<-c.fromLoop
		ap := c.getActivePoller(l)
		if ap.activeNodes == nil {
			ap.initNodes(l)
		}
		n := &ap.activeNodes[c.index]
		ap.currentNode = n
		t0 := cpu.TimeNow()
		ap.timeNow = t0
		p.LoopInput(l, n.looperOut)
		nVec := n.out.call(l, ap)
		ap.pollerStats.update(nVec, t0)
		l.pollerStats.update(nVec)
		c.toLoop <- struct{}{}
	}
}

func (l *Loop) doPollers() {
	{
		pending := l.nodeStateMain.advance(l.DataNodes)
		for _, p := range pending {
			n := l.DataNodes[p.nodeIndex].GetNode()
			if p.is_active {
				n.allocActivePoller(l)
			} else if !p.is_suspended {
				n.freeActivePoller(l)
			}
		}
	}

	for i := uint(0); i < l.activePollerPool.Len(); i++ {
		if l.activePollerPool.IsFree(i) {
			continue
		}
		n := l.activePollerPool.entries[i].pollerNode
		n.poller_elog(poller_elog_poll)
		// Start poller who will be blocked waiting on fromLoop.
		n.fromLoop <- struct{}{}
	}

	// Wait for pollers to finish.
	for i := uint(0); i < l.activePollerPool.Len(); i++ {
		if l.activePollerPool.IsFree(i) {
			continue
		}
		n := l.activePollerPool.entries[i].pollerNode
		<-n.toLoop
		n.poller_elog(poller_elog_poll_done)
	}

	if l.activePollerPool.Elts() == 0 {
		l.resetPollerStats()
	} else {
		l.doPollerStats()
	}
}

const (
	poller_elog_alloc = iota
	poller_elog_free
	poller_elog_add_pending
	poller_elog_event_wake
	poller_elog_poll
	poller_elog_poll_done
	poller_elog_suspended
	poller_elog_resumed
	poller_elog_data_activity
	poller_elog_suspend_activity
)

type poller_elog_kind uint32

func (k poller_elog_kind) String() string {
	switch k {
	case poller_elog_alloc:
		return "alloc"
	case poller_elog_free:
		return "free"
	case poller_elog_add_pending:
		return "add-pending"
	case poller_elog_event_wake:
		return "event-wake"
	case poller_elog_poll:
		return "poll"
	case poller_elog_poll_done:
		return "done"
	case poller_elog_suspended:
		return "suspended"
	case poller_elog_resumed:
		return "resumed"
	case poller_elog_data_activity:
		return "+data"
	case poller_elog_suspend_activity:
		return "+suspend"
	default:
		return fmt.Sprintf("unknown %d", int(k))
	}
}

type poller_elog struct {
	name  elog.StringRef
	kind  poller_elog_kind
	da, a int32
}

func (n *Node) poller_elog_i(kind poller_elog_kind, i uint) {
	e := poller_elog{
		name: n.elogNodeName,
		kind: kind,
		a:    int32(i),
	}
	elog.Add(&e)
}

func (n *Node) poller_elog_a(kind poller_elog_kind, da, a int32) {
	e := poller_elog{
		name: n.elogNodeName,
		kind: kind,
		da:   da,
		a:    a,
	}
	elog.Add(&e)
}

func (n *Node) poller_elog(kind poller_elog_kind) {
	e := poller_elog{
		name: n.elogNodeName,
		kind: kind,
	}
	elog.Add(&e)
}

func (e *poller_elog) Elog(l *elog.Log) {
	switch e.kind {
	case poller_elog_alloc, poller_elog_free:
		l.Logf("loop %s %s %d", e.kind, e.name, e.a)
	case poller_elog_data_activity, poller_elog_suspend_activity:
		l.Logf("loop %s %s %+d %d", e.kind, e.name, e.da, e.a)
	default:
		l.Logf("loop %s %s", e.kind, e.name)
	}
}
