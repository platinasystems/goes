// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loop

import (
	"github.com/platinasystems/go/elib/cpu"
	"github.com/platinasystems/go/elib/dep"
	"github.com/platinasystems/go/elib/elog"

	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type Node struct {
	l                       *Loop
	name                    string
	noder                   Noder
	index                   uint
	toLoop                  chan struct{}
	fromLoop                chan struct{}
	activePollerIndex       uint
	initOnce                sync.Once
	initWg                  sync.WaitGroup
	Next                    []string
	nextNodes               nextNodeVec
	nextIndexByNodeName     map[string]uint
	inputStats, outputStats nodeStats
	elogNodeName            elog.StringRef
	eventNode
	s nodeState
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
func (n *Node) GetLoop() *Loop           { return n.l }
func (n *Node) ThreadId() uint           { return n.activePollerIndex }
func nodeName(n Noder) string            { return n.GetNode().name }

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
	noders      []Noder
	noderByName map[string]Noder
	nodes       []*Node

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
	nodeStateMain
}

func (l *Loop) Seconds(t cpu.Time) float64 { return float64(t) * l.secsPerCycle }

func (l *Loop) startDataPoller(r inLooper) {
	n := r.GetNode()
	n.toLoop = make(chan struct{}, 1)
	n.fromLoop = make(chan struct{}, 1)
	go l.dataPoll(r)
}
func (l *Loop) startPollers() {
	if !poll_active {
		for _, n := range l.dataPollers {
			l.startDataPoller(n)
		}
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
	// Save elog if thread panics.
	defer func() {
		if elog.Enabled() {
			if err := recover(); err != nil {
				elog.Panic(fmt.Errorf("Run: %v", err))
				panic(err)
			}
		}
	}()

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

func (l *Loop) disableInterrupts(disable bool) {
	enable := !disable
	for _, n := range l.dataPollers {
		if x, ok := n.(InterruptEnabler); ok {
			x.InterruptEnable(enable)
		}
	}
	l.pollerStats.interruptsDisabled = disable
	elog.F1b("loop: irq disable %v", disable)
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
	n.index = uint(len(l.noders))
	n.activePollerIndex = ^uint(0)
	l.noders = append(l.noders, r)
	l.nodes = append(l.nodes, n)
	if l.noderByName == nil {
		l.noderByName = make(map[string]Noder)
	}
	if _, ok := l.noderByName[n.name]; ok {
		panic(fmt.Errorf("%s: more than one node with this name", n.name))
	}
	l.noderByName[n.name] = r
}

func isDataPoller(n Noder) (x inLooper, ok bool) { x, ok = n.(inLooper); return }

func (l *Loop) RegisterNode(n Noder, format string, args ...interface{}) {
	x := n.GetNode()
	x.name = fmt.Sprintf(format, args...)
	x.elogNodeName = elog.SetString(x.name)
	x.l = l
	for i := range x.Next {
		if _, err := l.AddNamedNext(n, x.Next[i]); err != nil {
			panic(err)
		}
	}

	start := l.registrationsNeedStart
	nOK := 0
	if h, ok := n.(EventHandler); ok {
		l.eventHandlers = append(l.eventHandlers, h)
		l.eventHandlerNodes = append(l.eventHandlerNodes, x)
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
		if q, ok := isDataPoller(d); ok {
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
