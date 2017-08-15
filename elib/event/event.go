// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package event

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cpu"

	"sync"
)

type Actor interface {
	EventAction()
	String() string
}

type TimedActor interface {
	Actor
	EventTime() cpu.Time
}

//go:generate gentemplate -d Package=event -id actor  -d VecType=ActorVec -d Type=Actor github.com/platinasystems/go/elib/vec.tmpl

func (p *timedEventPool) Compare(i, j int) int {
	ei, ej := p.events[i], p.events[j]
	return int(ei.EventTime() - ej.EventTime())
}

//go:generate gentemplate -d Package=event -id timedEvent -d PoolType=timedEventPool -d Type=TimedActor -d Data=events github.com/platinasystems/go/elib/pool.tmpl

type Pool struct {
	mu      sync.Mutex
	pool    timedEventPool
	fibheap elib.FibHeap
}

func (p *Pool) Elts() uint { return p.pool.Elts() }

func (p *Pool) Add(e TimedActor) (ei uint) {
	p.mu.Lock()
	defer p.mu.Unlock()
	ei = p.pool.GetIndex()
	p.pool.events[ei] = e
	p.fibheap.Add(ei)
	return ei
}

func (p *Pool) del(ei uint) {
	p.fibheap.Del(ei)
	p.pool.PutIndex(ei)
}

func (p *Pool) advance(t cpu.Time, iv *ActorVec) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for {
		ei, valid := p.fibheap.Min(&p.pool)
		if !valid {
			return
		}
		e := p.pool.events[ei]
		if e.EventTime() > t {
			break
		}
		p.del(ei)
		if iv != nil {
			*iv = append(*iv, e)
		} else {
			e.EventAction()
		}
	}
}

func (p *Pool) NextTime() (t cpu.Time, valid bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	ei, valid := p.fibheap.Min(&p.pool)
	if valid {
		t = p.pool.events[ei].EventTime()
	}
	return
}

func (p *Pool) Advance(t cpu.Time)                  { p.advance(t, nil) }
func (p *Pool) AdvanceAdd(t cpu.Time, iv *ActorVec) { p.advance(t, iv) }
