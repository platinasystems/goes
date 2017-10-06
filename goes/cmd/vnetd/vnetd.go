// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vnetd

import (
	"fmt"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
	"github.com/platinasystems/go/internal/redis/rpc/args"
	"github.com/platinasystems/go/internal/redis/rpc/reply"
	"github.com/platinasystems/go/internal/sockfile"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
)

const (
	Name    = "vnetd"
	Apropos = "FIXME"
	Usage   = "vnetd"
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() *Command { return new(Command) }

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

type Command struct {
	i Info
}

type Info struct {
	v         vnet.Vnet
	eventPool sync.Pool
	poller    ifStatsPoller
	pub       *publisher.Publisher
}

func (*Command) Apropos() lang.Alt { return apropos }
func (*Command) Kind() cmd.Kind    { return cmd.Daemon }
func (*Command) String() string    { return Name }
func (*Command) Usage() string     { return Usage }

func (c *Command) Main(...string) error {
	var (
		err error
		in  parse.Input
	)

	if err = redis.IsReady(); err != nil {
		return err
	}

	// never want to block vnet
	c.i.pub, err = publisher.New()
	if err != nil {
		return err
	}
	defer c.i.pub.Close()

	rpc.Register(&c.i)

	sock, err := sockfile.NewRpcServer(Name)
	if err != nil {
		return err
	}
	defer sock.Close()

	err = redis.Assign(redis.DefaultHash+":vnet.", Name, "Info")
	if err != nil {
		return err
	}

	c.i.eventPool.New = c.i.newEvent
	c.i.v.RegisterHwIfAddDelHook(c.i.hw_if_add_del)
	c.i.v.RegisterHwIfLinkUpDownHook(c.i.hw_if_link_up_down)
	c.i.v.RegisterSwIfAddDelHook(c.i.sw_if_add_del)
	c.i.v.RegisterSwIfAdminUpDownHook(c.i.sw_if_admin_up_down)
	if err = Hook(&c.i, &c.i.v); err != nil {
		return err
	}

	for GdbWait && gdb_wait == 0 {
		time.Sleep(100 * time.Millisecond)
	}

	sfn := sockfile.Path("vnet")
	in.SetString(fmt.Sprintf("cli { listen { no-prompt socket %s} }", sfn))
	go func(sfn string) {
		for {
			_, err := os.Stat(sfn)
			if err == nil {
				sockfile.Chgroup(sfn, "adm")
				break
			}
		}
	}(sfn)

	signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)

	err = c.i.v.Run(&in)
	CloseHook(&c.i, &c.i.v)
	closeDone <- err
	return nil
}

var closeDone = make(chan error)

func (c *Command) Close() (err error) {
	c.i.v.Quit()
	err = <-closeDone
	return
}

func (i *Info) Init() {
	i.poller.i = i
	i.poller.addEvent(0)
	i.poller.pollInterval = 5 // default 5 seconds
	i.initialPublish()
	i.set("ready", "true", true)
}

func (i *Info) Hset(args args.Hset, reply *reply.Hset) error {
	field := strings.TrimPrefix(args.Field, "vnet.")
	err := i.set(field, string(args.Value), false)
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

func (i *Info) sw_is_ok(si vnet.Si) bool {
	h := i.v.HwIferForSupSi(si)
	return h != nil && i.hw_is_ok(h.GetHwIf().Hi())
}

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
		itv    float64
		fec    ethernet.ErrorCorrectionType
	)
	if e.isReadyEvent {
		e.i.pub.Print("vnet.", e.key, ": ", e.value)
		return
	}
	e.in.Init(nil)
	e.in.Add(e.key, e.value)
	v := &e.i.v
	switch {
	case e.in.Parse("%v.speed %v", &hi, v, &bw):
		e.err <- hi.SetSpeed(v, bw)
	case e.in.Parse("%v.admin %v", &si, v, &enable):
		e.err <- si.SetAdminUp(v, bool(enable))
	case e.in.Parse("%v.media %s", &hi, v, &media):
		e.err <- hi.SetMedia(v, media)
	case e.in.Parse("%v.fec %v", &hi, v, &fec):
		e.err <- ethernet.SetInterfaceErrorCorrection(v, hi, fec)
	case e.in.Parse("pollInterval %f", &itv):
		if itv < 1 {
			e.err <- fmt.Errorf("pollInterval must be 1 second or longer")
		} else {
			e.i.poller.pollInterval = itv
			e.err <- nil
		}
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
		i.pub.Print("vnet.", key, ": ", value)
	}
	return
}

func (i *Info) initialPublish() {
	i.v.ForeachHwIf(UnixInterfacesOnly, func(hi vnet.Hi) {
		h := i.v.HwIf(hi)
		i.publish(hi.Name(&i.v)+".speed", h.Speed().String())
		i.publish(hi.Name(&i.v)+".media", h.Media())
	})
	i.publish("pollInterval", i.poller.pollInterval)
}

func (i *Info) publish(key string, value interface{}) {
	i.pub.Print("vnet.", key, ": ", value)
}

// One per each hw/sw interface from vnet.
type ifStatsPollerInterface struct {
	lastValues map[string]uint64
}

func (i *ifStatsPollerInterface) update(counter string, value uint64) (updated bool) {
	if i.lastValues == nil {
		i.lastValues = make(map[string]uint64)
	}
	if v, ok := i.lastValues[counter]; ok {
		if updated = v != value; updated {
			i.lastValues[counter] = value
		}
	} else {
		updated = true
		i.lastValues[counter] = value
	}
	return
}

//go:generate gentemplate -d Package=vnetd -id ifStatsPollerInterface -d VecType=ifStatsPollerInterfaceVec -d Type=ifStatsPollerInterface github.com/platinasystems/go/elib/vec.tmpl

type ifStatsPoller struct {
	vnet.Event
	i            *Info
	sequence     uint
	hwInterfaces ifStatsPollerInterfaceVec
	swInterfaces ifStatsPollerInterfaceVec
	pollInterval float64 // pollInterval in seconds
}

func (p *ifStatsPoller) publish(name, counter string, value uint64) {
	n := strings.Replace(counter, " ", "_", -1)
	p.i.publish(name+"."+n, value)
}
func (p *ifStatsPoller) addEvent(dt float64) { p.i.v.SignalEventAfter(p, dt) }
func (p *ifStatsPoller) String() string {
	return fmt.Sprintf("redis stats poller sequence %d", p.sequence)
}
func (p *ifStatsPoller) EventAction() {
	// Schedule next event in 5 seconds; do before fetching counters so that time interval is accurate.
	p.addEvent(p.pollInterval)

	start := time.Now()
	p.i.publish("poll.start", start.Format(time.StampMilli))
	// Publish all sw/hw interface counters even with zero values for first poll.
	// This was all possible counters have valid values in redis.
	// Otherwise only publish to redis when counter values change.
	includeZeroCounters := p.sequence == 0
	p.i.v.ForeachHwIfCounter(includeZeroCounters, UnixInterfacesOnly,
		func(hi vnet.Hi, counter string, value uint64) {
			p.hwInterfaces.Validate(uint(hi))
			if p.hwInterfaces[hi].update(counter, value) {
				p.publish(hi.Name(&p.i.v), counter, value)
			}
		})
	p.i.v.ForeachSwIfCounter(includeZeroCounters,
		func(si vnet.Si, counter string, value uint64) {
			p.swInterfaces.Validate(uint(si))
			if p.swInterfaces[si].update(counter, value) {
				p.publish(si.Name(&p.i.v), counter, value)
			}
		})
	stop := time.Now()
	p.i.publish("poll.finish", stop.Format(time.StampMilli))

	p.sequence++
}
