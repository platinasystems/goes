// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vnetd

import (
	"fmt"
	"net/rpc"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
	"github.com/platinasystems/go/internal/redis/rpc/args"
	"github.com/platinasystems/go/internal/redis/rpc/reply"
	"github.com/platinasystems/go/internal/sockfile"
	"github.com/platinasystems/go/vnet"
)

const Name = "vnetd"

// Enable publish of Non-unix (e.g. non-tuntap) interfaces.
// This will include all vnet interfaces.
var UnixInterfacesOnly bool

// Wait for gdb before starting vnet.
var GdbWait bool

// In gdb issue command "p 'vnetd.gdb_wait'=1" to break out of loop and
// start vnet.
var gdb_wait int

// Machines may reassign this for platform sepecific init before vnet.Run.
var Hook = func(*Info, *vnet.Vnet) error { return nil }

// Machines may reassign this for platform sepecific cleanup after vnet.Quit.
var CloseHook = func(*Info, *vnet.Vnet) error { return nil }

var Prefixes = []string{"eth-"}

type cmd struct {
	i Info
}

type Info struct {
	v         vnet.Vnet
	eventPool sync.Pool
	poller    ifStatsPoller
	pub       *publisher.Publisher
}

func New() *cmd { return &cmd{} }

func (*cmd) Kind() goes.Kind { return goes.Daemon }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return Name }

func (cmd *cmd) Main(...string) error {
	var (
		err error
		in  parse.Input
	)

	// never want to block vnet
	cmd.i.pub, err = publisher.New()
	if err != nil {
		return err
	}
	defer cmd.i.pub.Close()

	rpc.Register(&cmd.i)

	sock, err := sockfile.NewRpcServer(Name)
	if err != nil {
		return err
	}
	defer sock.Close()

	for _, prefix := range Prefixes {
		key := fmt.Sprintf("%s:%s", redis.DefaultHash, prefix)
		err = redis.Assign(key, Name, "Info")
		if err != nil {
			return err
		}
	}

	cmd.i.eventPool.New = cmd.i.newEvent
	cmd.i.v.RegisterHwIfAddDelHook(cmd.i.hw_if_add_del)
	cmd.i.v.RegisterHwIfLinkUpDownHook(cmd.i.hw_if_link_up_down)
	cmd.i.v.RegisterSwIfAddDelHook(cmd.i.sw_if_add_del)
	cmd.i.v.RegisterSwIfAdminUpDownHook(cmd.i.sw_if_admin_up_down)
	if err = Hook(&cmd.i, &cmd.i.v); err != nil {
		return err
	}

	for GdbWait && gdb_wait == 0 {
		time.Sleep(100 * time.Millisecond)
	}

	sfn := sockfile.Path("vnet")
	in.SetString(fmt.Sprintf("cli { listen { no-prompt socket %s} }", sfn))
	go sockfile.Chgroup(sfn, "adm")

	return cmd.i.v.Run(&in)
}

func (cmd *cmd) Close() error {
	// FIXME the following isn't working.
	// cmd.i.v.Quit()
	// return CloseHook(&cmd.i, &cmd.i.v)
	return nil
}

func Init(i *Info) {
	i.poller.i = i
	i.poller.addEvent(0)
	i.initialPublish()
	i.set("vnet.ready", "true", true)
}

func (i *Info) Hset(args args.Hset, reply *reply.Hset) error {
	err := i.set(args.Field, string(args.Value), false)
	if err == nil {
		*reply = 1
	}
	return err
}

func (i *Info) hw_is_ok(hi vnet.Hi) bool {
	h := i.v.HwIfer(hi)
	hw := i.v.HwIf(hi)
	if !hw.IsProvisioned() {
		return false
	}
	return !UnixInterfacesOnly || h.IsUnix()
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
	i.publish(hi.Name(&i.v)+".link", parse.Enable(isUp))
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
	i            *Info
	in           parse.Input
	key, value   string
	err          chan error
	isReadyEvent bool
}

func (i *Info) newEvent() interface{} {
	return &event{
		i:   i,
		err: make(chan error, 1),
	}
}

func (e *event) String() string {
	return fmt.Sprintf("redis set %s = %s", e.key, e.value)
}

func (e *event) EventAction() {
	var (
		hi     vnet.Hi
		si     vnet.Si
		bw     vnet.Bandwidth
		enable parse.Enable
		media  string
	)
	if e.isReadyEvent {
		e.i.pub.Print(e.key, ": ", e.value)
		return
	}
	e.in.Init(nil)
	e.in.Add(e.key, e.value)
	switch {
	case e.in.Parse("%v.speed %v", &hi, &e.i.v, &bw):
		e.err <- hi.SetSpeed(&e.i.v, bw)
	case e.in.Parse("%v.admin %v", &si, &e.i.v, &enable):
		e.err <- si.SetAdminUp(&e.i.v, bool(enable))
	case e.in.Parse("%v.media %s", &hi, &e.i.v, &media):
		e.err <- hi.SetMedia(&e.i.v, media)
	default:
		e.err <- fmt.Errorf("can't set %s to %v", e.key, e.value)
	}
	e.i.eventPool.Put(e)
}

func (i *Info) set(key, value string, isReadyEvent bool) (err error) {
	e := i.eventPool.Get().(*event)
	e.key = key
	e.value = value
	e.isReadyEvent = isReadyEvent
	i.v.SignalEvent(e)
	if isReadyEvent {
		return
	}
	if err = <-e.err; err == nil {
		i.pub.Print(key, ": ", value)
	}
	return
}

func (i *Info) initialPublish() {
	i.v.ForeachHwIf(UnixInterfacesOnly, func(hi vnet.Hi) {
		h := i.v.HwIf(hi)
		i.publish(hi.Name(&i.v)+".speed", h.Speed().String())
	})
}

func (i *Info) publish(key string, value interface{}) {
	i.pub.Print(key, ": ", value)
}

type ifStatsPoller struct {
	vnet.Event
	i            *Info
	sequence     uint
	hwLastValues elib.Uint64Vec
	swLastValues elib.Uint64Vec
}

func (p *ifStatsPoller) publish(name, counter string, value uint64) {
	n := strings.Replace(counter, " ", "_", -1)
	p.i.publish(name+"."+n, value)
}
func (p *ifStatsPoller) addEvent(dt float64) { p.i.v.AddTimedEvent(p, dt) }
func (p *ifStatsPoller) String() string {
	return fmt.Sprintf("redis stats poller sequence %d", p.sequence)
}
func (p *ifStatsPoller) EventAction() {
	// Schedule next event in 5 seconds; do before fetching counters so that time interval is accurate.
	p.addEvent(5)

	// Publish all sw/hw interface counters even with zero values for first poll.
	// This was all possible counters have valid values in redis.
	// Otherwise only publish to redis when counter values change.
	includeZeroCounters := p.sequence == 0
	p.i.v.ForeachHwIfCounter(includeZeroCounters, UnixInterfacesOnly,
		func(hi vnet.Hi, counter string, value uint64) {
			p.hwLastValues.Validate(uint(hi))
			if v := p.hwLastValues[hi]; v != value || includeZeroCounters {
				p.hwLastValues[hi] = value
				p.publish(hi.Name(&p.i.v), counter, value)
			}
		})
	p.i.v.ForeachSwIfCounter(includeZeroCounters,
		func(si vnet.Si, counter string, value uint64) {
			p.swLastValues.Validate(uint(si))
			if v := p.swLastValues[si]; v != value || includeZeroCounters {
				p.swLastValues[si] = value
				p.publish(si.Name(&p.i.v), counter, value)
			}
		})

	p.sequence++
}
