// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loop

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cpu"
	"github.com/platinasystems/go/elib/dep"
	"github.com/platinasystems/go/elib/elog"

	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Node struct {
	name                    string
	noder                   Noder
	index                   uint
	loop                    *Loop
	toLoop                  chan struct{}
	fromLoop                chan struct{}
	node_flags              node_flags
	activePollerIndex       uint
	initOnce                sync.Once
	initWg                  sync.WaitGroup
	Next                    []string
	nextNodes               nextNodeVec
	nextIndexByNodeName     map[string]uint
	inputStats, outputStats nodeStats
	elogNodeName            elog.StringRef
	eventNode
}

type node_flags uint32

func (n *Node) get_flags() node_flags { return node_flags(atomic.LoadUint32((*uint32)(&n.node_flags))) }

const (
	log2_node_active, node_active node_flags = iota, 1 << iota
	log2_node_suspended, node_suspended
	log2_node_resumed, node_resumed
	log2_node_polling, node_polling
)

var node_flag_strings = [...]string{
	log2_node_active:    "active",
	log2_node_suspended: "suspended",
	log2_node_resumed:   "resumed",
	log2_node_polling:   "polling",
}

func (x node_flags) String() string { return elib.FlagStringer(node_flag_strings[:], elib.Word(x)) }

func (n *Node) is_active() bool    { return n.get_flags()&node_active != 0 }
func (n *Node) is_polling() bool   { return n.get_flags()&node_polling != 0 }
func (n *Node) is_suspended() bool { return n.get_flags()&node_suspended != 0 }
func (n *Node) is_resumed() bool   { return n.get_flags()&node_resumed != 0 }

func (n *Node) set_flag(f node_flags, v bool) (new node_flags) {
	for {
		old := n.get_flags()
		new = old
		if v {
			new |= f
		} else {
			new &^= f
		}
		if n.node_flags.compare_and_swap(old, new) {
			break
		}
	}
	return
}

func (n *Node) set_active(v bool) { n.set_flag(node_active, v) }

func (f *node_flags) compare_and_swap(old, new node_flags) (swapped bool) {
	return atomic.CompareAndSwapUint32((*uint32)(f), uint32(old), uint32(new))
}

type nextNode struct {
	name      string
	nodeIndex uint
	in        LooperIn
}

//go:generate gentemplate -d Package=loop -id nextNode -d VecType=nextNodeVec -d Type=nextNode github.com/platinasystems/go/elib/vec.tmpl

func (n *Node) GetNode() *Node           { return n }
func (n *Node) Index() uint              { return n.index }
func (n *Node) Name() string             { return n.name }
func (n *Node) ElogName() elog.StringRef { return n.elogNodeName }
func (n *Node) GetLoop() *Loop           { return n.loop }
func (n *Node) ThreadId() uint           { return n.activePollerIndex }
func nodeName(n Noder) string            { return n.GetNode().name }

func (n *Node) getActivePoller(l *Loop) *activePoller {
	return l.activePollerPool.entries[n.activePollerIndex]
}

func (n *Node) allocActivePoller(l *Loop) {
	i := l.activePollerPool.GetIndex()
	a := l.activePollerPool.entries[i]
	if a == nil {
		a = &activePoller{}
		l.activePollerPool.entries[i] = a
	}
	a.index = uint16(i)
	n.activePollerIndex = i
	a.pollerNode = n
}

func (a *activePoller) flushNodeStats(l *Loop) {
	for i := range a.activeNodes {
		ani := &a.activeNodes[i]
		ni := l.DataNodes[ani.index].GetNode()

		ni.inputStats.current.add_raw(&ani.inputStats)
		ani.inputStats.zero()

		ni.outputStats.current.add_raw(&ani.outputStats)
		ani.outputStats.zero()
	}
}

func (n *Node) freeActivePoller(l *Loop) {
	a := n.getActivePoller(l)
	a.flushNodeStats(l)
	a.pollerNode = nil
	i := n.activePollerIndex
	l.activePollerPool.PutIndex(i)
	n.activePollerIndex = ^uint(0)
}

func (n *Node) Activate(enable bool) (was bool) {
	for {
		old := n.get_flags()
		was = old&node_active != 0
		if was == enable {
			break
		}
		new := old
		if enable {
			new |= node_active
		} else {
			new &^= node_active
		}
		if n.node_flags.compare_and_swap(old, new) {
			n.pollerElog(poller_activate, new)
			break
		}
	}
	if was != enable {
		n.changeActive(enable)
	}
	return
}

func (n *Node) changeActive(enable bool) {
	l := n.loop
	if _, eventWait := l.activePollerState.changeActive(enable); eventWait {
		l.Interrupt()
	}
}

type activateEvent struct{ n *Node }

func (e *activateEvent) EventAction()   { e.n.Activate(true) }
func (e *activateEvent) String() string { return fmt.Sprintf("activate %s", e.n.name) }

func (n *Node) ActivateAfterTime(dt float64) {
	if was := n.Activate(false); was {
		n.activateEvent.n = n
		le := n.loop.getLoopEvent(&n.activateEvent, elog.PointerToFirstArg(&n))
		n.loop.addTimedEvent(le, dt)
	}
}

type Noder interface {
	GetNode() *Node
}

type Initer interface {
	Noder
	LoopInit(l *Loop)
}

type Exiter interface {
	LoopExit(l *Loop)
}

type Loop struct {
	DataNodes      []Noder
	dataNodeByName map[string]Noder

	loopIniters []Initer
	loopExiters []Exiter

	dataPollers       []inLooper
	activePollerState activePollerState
	activePollerPool  activePollerPool
	pollerStats       pollerStats

	wg sync.WaitGroup

	registrationsNeedStart bool
	initialNodesRegistered bool
	startTime              cpu.Time
	now                    cpu.Time
	cyclesPerSec           float64
	secsPerCycle           float64
	timeDurationPerCycle   float64
	timeLastRuntimeClear   time.Time

	Cli LoopCli
	Config
	eventMain
	loggerMain
}

func (l *Loop) Seconds(t cpu.Time) float64 { return float64(t) * l.secsPerCycle }

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

func (l *Loop) startPollers() {
	for _, n := range l.dataPollers {
		l.startDataPoller(n)
	}
}

func (l *Loop) Suspend(in *In) (resumed bool) {
	a := l.activePollerPool.entries[in.activeIndex]
	p := a.pollerNode
	for {
		old := p.get_flags()
		if resumed = old&node_resumed != 0; resumed {
			p.pollerElog(poller_abort_suspend, old)
			break
		}
		new := old
		new &^= node_resumed
		new |= node_suspended
		if p.node_flags.compare_and_swap(old, new) {
			p.pollerElog(poller_suspend, new)
			break
		}
	}
	if !resumed {
		p.inputStats.current.suspends++

		// Signal polling done to main loop.
		p.toLoop <- struct{}{}
		// Wait for continue (resume) signal from main loop.
		<-p.fromLoop
	}
	new := p.set_flag(node_resumed|node_suspended, false)
	p.pollerElog(poller_resumed, new)
	return
}

func (l *Loop) Resume(in *In) {
	a := l.activePollerPool.entries[in.activeIndex]
	if p := a.pollerNode; p != nil {
		for {
			old := p.get_flags()
			if old&(node_active|node_suspended) == 0 {
				p.pollerElog(poller_resume, old)
				return
			}
			new := old
			new |= node_active
			new |= node_resumed
			new &^= node_suspended
			if p.node_flags.compare_and_swap(old, new) {
				p.pollerElog(poller_resume, new)
				if old&node_active == 0 {
					p.changeActive(true)
				}
				break
			}
		}
	}
}

func (l *Loop) dataPoll(p inLooper) {
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

func (l *Loop) startDataPoller(n inLooper) {
	c := n.GetNode()
	c.toLoop = make(chan struct{}, 1)
	c.fromLoop = make(chan struct{}, 1)
	go l.dataPoll(n)
}

func (l *Loop) doPollers() {
	for _, p := range l.dataPollers {
		n := p.GetNode()
		if !n.is_active() || n.is_suspended() {
			continue
		}
		if n.activePollerIndex == ^uint(0) {
			n.allocActivePoller(n.loop)
		}
		n.set_flag(node_polling, true)
		n.pollerElog(poller_wake, n.get_flags())
		// Start poller who will be blocked waiting on fromLoop.
		n.fromLoop <- struct{}{}
	}

	// Wait for pollers to finish.
	nFreed := uint(0)
	for i := uint(0); i < l.activePollerPool.Len(); i++ {
		if l.activePollerPool.IsFree(i) {
			continue
		}
		n := l.activePollerPool.entries[i].pollerNode
		if !n.is_polling() {
			continue
		}

		<-n.toLoop
		n.set_flag(node_polling, false)
		n.pollerElog(poller_wait, n.get_flags())

		// If not active anymore we can free it now.
		// TODO: smp races.  Disabled for now.
		if false && !(n.is_active() || n.is_suspended()) {
			nFreed++
			if !l.activePollerPool.IsFree(n.activePollerIndex) {
				n.freeActivePoller(l)
			}
		}
	}

	// atomic.AddUint32(&l.nActivePollers, -uint32(nFreed))
	if false {
		l.resetPollerStats()
	} else {
		l.doPollerStats()
	}
}

func (l *Loop) timerInit() {
	t := cpu.Time(0)
	t.Cycles(1 * cpu.Second)
	l.cyclesPerSec = float64(t)
	l.secsPerCycle = 1 / l.cyclesPerSec
	l.timeDurationPerCycle = l.secsPerCycle * float64(time.Second)
}

func (l *Loop) TimeDiff(t0, t1 cpu.Time) float64 { return float64(t1-t0) * l.secsPerCycle }

type initHook func(l *Loop)

//go:generate gentemplate -id initHook -d Package=loop -d DepsType=initHookVec -d Type=initHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl

var initHooks, exitHooks initHookVec

func AddInit(f initHook, d ...*dep.Dep) { initHooks.Add(f, d...) }
func AddExit(f initHook, d ...*dep.Dep) { exitHooks.Add(f, d...) }

func (l *Loop) callInitHooks() {
	for i := range initHooks.hooks {
		initHooks.Get(i)(l)
	}
}

func (l *Loop) callExitHooks() {
	for i := range exitHooks.hooks {
		exitHooks.Get(i)(l)
	}
}

func (l *Loop) callInitNode(n Initer, isCall bool) {
	c := n.GetNode()
	wg := &l.wg
	if isCall {
		wg = &c.initWg
	}
	c.initOnce.Do(func() {
		wg.Add(1)
		go func() {
			n.LoopInit(l)
			wg.Done()
		}()
	})
}
func (l *Loop) CallInitNode(n Initer)  { l.callInitNode(n, true) }
func (l *Loop) startInitNode(n Initer) { l.callInitNode(n, false) }

func (l *Loop) doInitNodes() {
	for _, i := range l.loopIniters {
		l.startInitNode(i)
	}
	l.wg.Wait()
}

func (l *Loop) doExit() {
	l.callExitHooks()
	for i := range l.loopExiters {
		l.loopExiters[i].LoopExit(l)
	}
}

type Config struct {
	LogWriter         io.Writer
	QuitAfterDuration float64
}

type loopQuit struct {
	l        *Loop
	duration float64
	verbose  bool
}

func (l *loopQuit) String() string { return "quit" }
func (l *loopQuit) EventAction() {
	if l.verbose {
		l.l.Logln("quitting after", l.duration)
	}
	l.l.Quit()
}

func (l *Loop) quitAfter() {
	e := &loopQuit{l: l, verbose: false, duration: l.QuitAfterDuration}
	f := l.getLoopEvent(e, elog.PointerToFirstArg(&l))
	l.addTimedEvent(f, l.QuitAfterDuration)
}

func (l *Loop) Run() {
	elog.Enable(true)
	go elog.PrintOnHangupSignal(os.Stderr, false)

	l.timerInit()
	l.startTime = cpu.TimeNow()
	l.timeLastRuntimeClear = time.Now()
	l.cliInit()
	l.eventMain.Init(l)
	l.startPollers()
	l.registrationsNeedStart = true
	l.callInitHooks()
	l.doInitNodes()
	// Now that all initial nodes have been registered, initialize node graph.
	l.graphInit()
	if l.QuitAfterDuration > 0 {
		l.quitAfter()
	}
	for {
		if quit := l.doEvents(); quit {
			break
		}
		l.doPollers()
	}
	l.doExit()
}

type pollerCounts struct {
	nActiveNodes   uint32
	nActiveVectors uint32
}

type pollerStats struct {
	loopCount          uint64
	updateCount        uint64
	current            pollerCounts
	history            [1 << log2PollerHistorySize]pollerCounts
	interruptsDisabled bool
}

const (
	log2LoopsPerStatsUpdate = 7
	loopsPerStatsUpdate     = 1 << log2LoopsPerStatsUpdate
	log2PollerHistorySize   = 1
	// When vector rate crosses threshold disable interrupts and switch to polling mode.
	interruptDisableThreshold float64 = 10
)

type InterruptEnabler interface {
	InterruptEnable(enable bool)
}

func (l *Loop) resetPollerStats() {
	s := &l.pollerStats
	s.loopCount = 0
	for i := range s.history {
		s.history[i].reset()
	}
	s.current.reset()
	if s.interruptsDisabled {
		l.disableInterrupts(false)
	}
}

func (l *Loop) disableInterrupts(disable bool) {
	enable := !disable
	for _, n := range l.dataPollers {
		if x, ok := n.(InterruptEnabler); ok {
			x.InterruptEnable(enable)
			n.GetNode().Activate(disable)
		}
	}
	l.pollerStats.interruptsDisabled = disable
	elog.F1b("loop: irq disable %v", disable)
}

func (l *Loop) doPollerStats() {
	s := &l.pollerStats
	s.loopCount++
	if s.loopCount&(1<<log2LoopsPerStatsUpdate-1) == 0 {
		s.history[s.updateCount&(1<<log2PollerHistorySize-1)] = s.current
		s.updateCount++
		disable := s.current.vectorRate() > interruptDisableThreshold
		if disable != s.interruptsDisabled {
			l.disableInterrupts(disable)
		}
		s.current.reset()
	}
}

func (s *pollerStats) update(nVec uint) {
	v := uint32(0)
	if nVec > 0 {
		v = 1
	}
	c := &s.current
	atomic.AddUint32(&c.nActiveVectors, uint32(nVec))
	atomic.AddUint32(&c.nActiveNodes, v)
}

func (c *pollerCounts) vectorRate() float64 {
	return float64(c.nActiveVectors) / float64(1<<log2LoopsPerStatsUpdate)
}

func (c *pollerCounts) reset() {
	c.nActiveVectors = 0
	c.nActiveNodes = 0
}

func (s *pollerStats) VectorRate() float64 {
	return s.history[(s.updateCount-1)&(1<<log2PollerHistorySize-1)].vectorRate()
}

func (l *Loop) addDataNode(r Noder) {
	n := r.GetNode()
	n.noder = r
	n.index = uint(len(l.DataNodes))
	n.activePollerIndex = ^uint(0)
	l.DataNodes = append(l.DataNodes, r)
	if l.dataNodeByName == nil {
		l.dataNodeByName = make(map[string]Noder)
	}
	if _, ok := l.dataNodeByName[n.name]; ok {
		panic(fmt.Errorf("%s: more than one node with this name", n.name))
	}
	l.dataNodeByName[n.name] = r
}

func (l *Loop) RegisterNode(n Noder, format string, args ...interface{}) {
	x := n.GetNode()
	x.name = fmt.Sprintf(format, args...)
	x.elogNodeName = elog.SetString(x.name)
	x.loop = l
	for i := range x.Next {
		if _, err := l.AddNamedNext(n, x.Next[i]); err != nil {
			panic(err)
		}
	}

	start := l.registrationsNeedStart
	nOK := 0
	if h, ok := n.(EventHandler); ok {
		l.eventMain.handlers = append(l.eventMain.handlers, h)
		if start {
			l.startHandler(h)
		}
		nOK++
	}
	if d, isOut := n.(outNoder); isOut {
		nok := 0
		if _, ok := d.(inOutLooper); ok {
			nok++
		}
		if q, ok := d.(inLooper); ok {
			l.dataPollers = append(l.dataPollers, q)
			if start {
				l.startDataPoller(q)
			}
			nok++
		}
		if nok == 0 {
			// Accept output only node.
			nok = 1
		}
		l.addDataNode(n)
		nOK += nok
	} else if _, isIn := n.(inNoder); isIn {
		if _, ok := n.(outLooper); ok {
			l.addDataNode(n)
			nOK += 1
		} else {
			panic(fmt.Errorf("%s: missing LoopOutput method", x.name))
		}
	}
	if p, ok := n.(Initer); ok {
		l.loopIniters = append(l.loopIniters, p)
		if start {
			l.startInitNode(p)
		}
		nOK++
	}
	if p, ok := n.(Exiter); ok {
		l.loopExiters = append(l.loopExiters, p)
		nOK++
	}
	if nOK == 0 {
		panic(fmt.Errorf("unkown node type: %T", n))
	}
}
