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

type Config struct {
	// Enable publish of Non-unix (e.g. non-tuntap) interfaces.  This will include all vnet interfaces.
	UnixInterfacesOnly bool
	// Publish all counters including those with zero values.
	PublishAllCounters bool
	// Wait for gdb before starting vnet.
	GdbWait bool
}

type Info struct {
	v         *vnet.Vnet
	prefix    string
	prefixes  []string
	eventPool sync.Pool
	Config
	ifStatsPoller
}

func New(v *vnet.Vnet, cf Config) (i *Info) {
	const prefix = "vnet."
	i = &Info{
		v:        v,
		prefix:   prefix,
		prefixes: []string{prefix},
		Config:   cf,
	}
	i.eventPool.New = i.newEvent
	v.RegisterHwIfAddDelHook(i.hw_if_add_del)
	v.RegisterHwIfLinkUpDownHook(i.hw_if_link_up_down)
	v.RegisterSwIfAddDelHook(i.sw_if_add_del)
	v.RegisterSwIfAdminUpDownHook(i.sw_if_admin_up_down)
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

func (i *Info) hw_is_ok(hi vnet.Hi) bool {
	h := i.v.HwIfer(hi)
	hw := i.v.HwIf(hi)
	if !hw.IsProvisioned() {
		return false
	}
	return !i.UnixInterfacesOnly || h.IsUnix()
}

func (i *Info) sw_is_ok(si vnet.Si) bool { return i.hw_is_ok(i.v.SupHi(si)) }

func (i *Info) sw_if_add_del(v *vnet.Vnet, si vnet.Si, isDel bool) (err error) {
	i.sw_if_admin_up_down(v, si, false)
	return
}

func (i *Info) sw_if_admin_up_down(v *vnet.Vnet, si vnet.Si, isUp bool) (err error) {
	if i.sw_is_ok(si) {
		i.publish(si.Name(v)+".admin", parse.Enable(isUp))
	}
	return
}

func (i *Info) publish_link(hi vnet.Hi, isUp bool) {
	i.publish(hi.Name(i.v)+".link", parse.Enable(isUp))
}

func (i *Info) hw_if_add_del(v *vnet.Vnet, hi vnet.Hi, isDel bool) (err error) {
	i.hw_if_link_up_down(v, hi, false)
	return
}

func (i *Info) hw_if_link_up_down(v *vnet.Vnet, hi vnet.Hi, isUp bool) (err error) {
	if i.hw_is_ok(hi) {
		i.publish_link(hi, isUp)
	}
	return
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
		hi     vnet.Hi
		si     vnet.Si
		bw     vnet.Bandwidth
		enable parse.Enable
	)
	e.in.Init(nil)
	e.in.Add(e.key, e.value)
	switch {
	case e.in.Parse(e.i.prefix+"%v.speed %v", &hi, e.i.v, &bw):
		e.err <- hi.SetSpeed(e.i.v, bw)
	case e.in.Parse(e.i.prefix+"%v.admin %v", &si, e.i.v, &enable):
		e.err <- si.SetAdminUp(e.i.v, bool(enable))
	default:
		e.err <- info.CantSet(e.key)
	}
	e.i.eventPool.Put(e)
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

func (i *Info) gdbWait() {
	// In gdb issue command "p 'vnetinfo.gdb_wait'=1" to break out of loop and start vnet.
	for i.GdbWait && gdb_wait == 0 {
		time.Sleep(100 * time.Millisecond)
	}
}

func (i *Info) Start() error {
	i.gdbWait()
	var in parse.Input
	in.SetString("cli { listen { no-prompt socket " + vnetCmdSock + "} }")
	return i.v.Run(&in)
}

func (i *Info) initialPublish() {
	i.v.ForeachHwIf(i.UnixInterfacesOnly, func(hi vnet.Hi) {
		h := i.v.HwIf(hi)
		i.publish(hi.Name(i.v)+".speed", h.Speed().String())
	})
}

func (i *Info) Init() {
	p := &i.ifStatsPoller
	p.i = i
	p.addEvent(0)
	i.initialPublish()
}

type ifStatsPoller struct {
	vnet.Event
	i        *Info
	sequence uint
}

func (i *Info) publish(key string, value interface{}) { info.Publish(i.prefix+key, value) }

func (p *ifStatsPoller) publish(name, counter string, value uint64) {
	n := strings.Replace(counter, " ", "_", -1)
	p.i.publish(name+"."+n, value)
}
func (p *ifStatsPoller) addEvent(dt float64) { p.i.v.AddTimedEvent(p, dt) }
func (p *ifStatsPoller) String() string      { return "redis stats poller" }
func (p *ifStatsPoller) EventAction() {
	// Enable to represent all possible counters in redis (most with 0 values)
	includeZeroCounters := p.sequence == 0 && p.i.PublishAllCounters
	p.i.v.ForeachHwIfCounter(includeZeroCounters, p.i.UnixInterfacesOnly,
		func(hi vnet.Hi, counter string, value uint64) {
			p.publish(hi.Name(p.i.v), counter, value)
		})
	p.i.v.ForeachSwIfCounter(includeZeroCounters,
		func(si vnet.Si, counter string, value uint64) {
			p.publish(si.Name(p.i.v), counter, value)
		})
	p.addEvent(5) // schedule next event in 5 seconds
	p.sequence++
}
