// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package vnetinfo

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/machined/info"
	"github.com/platinasystems/go/vnet"

	"fmt"
	"strings"
	"sync"
	"time"
)

const Name = "vnet"

type Info struct {
	v         *vnet.Vnet
	prefixes  []string
	eventPool sync.Pool
	statsPoller
}

func NewInfo(v *vnet.Vnet) (i *Info) {
	i = &Info{
		v:        v,
		prefixes: []string{"eth-"},
	}
	i.eventPool.New = i.newEvent
	return
}

func (*Info) String() string { return Name }

func (*Info) Main(...string) error { return nil }

func (i *Info) Close() (err error) {
	// Exit vnet main loop.
	i.v.Quit()
	return
}

func (*Info) Del(key string) error { return info.CantDel(key) }

func (i *Info) Prefixes(p ...string) []string {
	if len(p) > 0 {
		i.prefixes = p
	}
	return i.prefixes
}

type event struct {
	vnet.Event
	i          *Info
	in         parse.Input
	key, value string
	err        chan error
}

func (i *Info) newEvent() interface{} {
	return &event{
		i:   i,
		err: make(chan error, 1),
	}
}

func (e *event) String() string { return fmt.Sprintf("redis set %s = %s", e.key, e.value) }
func (e *event) EventAction() {
	var (
		hi vnet.Hi
		bw vnet.Bandwidth
	)
	e.in.Init(nil)
	e.in.Add(e.key, e.value)
	switch {
	case e.in.Parse("%v.speed %v", &hi, e.i.v, &bw):
		e.err <- hi.SetSpeed(e.i.v, bw)
	default:
		e.err <- info.CantSet(e.key)
	}
	e.i.eventPool.Put(e)
}

func (i *Info) initialPublish() {
	i.v.ForeachHwIf(unixInterfacesOnly, func(hi vnet.Hi) {
		h := i.v.HwIf(hi)
		info.Publish(hi.Name(i.v)+".speed", h.Speed().String())
	})
}

func (i *Info) Set(key, value string) (err error) {
	e := i.eventPool.Get().(*event)
	e.key = key
	e.value = value
	i.v.SignalEvent(e)
	if err = <-e.err; err == nil {
		info.Publish(key, value)
	}
	return
}

var gdb_wait int

func gdbWait() {
	// Change false to true to enable.
	// In gdb issue command "p 'vnetinfo.gdb_wait'=1" to break out of loop and start vnet.
	const enable = false
	for enable && gdb_wait == 0 {
		time.Sleep(100 * time.Millisecond)
	}
}

func (i *Info) Start() error {
	gdbWait()
	i.statsPoller.i = i
	var in parse.Input
	in.Add("cli { listen { no-prompt socket " + vnetCmdSock + "} }")
	return i.v.Run(&in)
}

const unixInterfacesOnly = true // only front panel ports (e.g. no bcm-cpu or loopback ports)

func (i *Info) Init() {
	i.initialPublish()
	i.statsPoller.addEvent(0)
}

type statsPoller struct {
	vnet.Event
	i        *Info
	sequence uint
}

func publishIfCounter(name, counter string, value uint64) {
	n := strings.Replace(counter, " ", "_", -1)
	info.Publish(name+"."+n, value)
}
func (p *statsPoller) addEvent(dt float64) { p.i.v.AddTimedEvent(p, dt) }
func (p *statsPoller) String() string      { return "redis stats poller" }
func (p *statsPoller) EventAction() {
	// Enable to represent all possible counters in redis (most with 0 values)
	includeZeroCounters := p.sequence == 0 && false
	p.i.v.ForeachHwIfCounter(includeZeroCounters, unixInterfacesOnly,
		func(hi vnet.Hi, counter string, value uint64) {
			publishIfCounter(hi.Name(p.i.v), counter, value)
		})
	p.i.v.ForeachSwIfCounter(includeZeroCounters,
		func(si vnet.Si, counter string, value uint64) {
			publishIfCounter(si.Name(p.i.v), "vnet "+counter, value)
		})
	p.addEvent(5) // schedule next event in 5 seconds
	p.sequence++
}
