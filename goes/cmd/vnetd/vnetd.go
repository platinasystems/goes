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

	"github.com/platinasystems/atsock"
	"github.com/platinasystems/elib/parse"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/redis"
	"github.com/platinasystems/redis/publisher"
	"github.com/platinasystems/redis/rpc/args"
	"github.com/platinasystems/redis/rpc/reply"
	"github.com/platinasystems/vnet"
	"github.com/platinasystems/vnet/ethernet"
	"github.com/platinasystems/xeth"
	// "github.com/confluentinc/confluent-kafka-go/kafka"
	"regexp"
)

// Enable publish of Non-unix (e.g. non-tuntap) interfaces.
// This will include all vnet interfaces.
var UnixInterfacesOnly bool

// Machines may reassign this for platform sepecific init before vnet.Run.
var Hook = func(func(), *vnet.Vnet) error { return nil }

// Machines may reassign this for platform sepecific cleanup after vnet.Quit.
var CloseHook = func(*Info, *vnet.Vnet) error { return nil }

var Counter = func(s string) string { return s }

type Command struct {
	Init func()
	init sync.Once

	i Info
}

type Info struct {
	v         vnet.Vnet
	eventPool sync.Pool
	poller    ifStatsPoller
	fastPoller    fastIfStatsPoller
	pub       *publisher.Publisher
	// producer	*kafka.Producer
}

func (*Command) String() string { return "vnetd" }

func (*Command) Usage() string { return "vnetd" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "FIXME",
	}
}

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

const chanDepth = 1 << 16

var closeDone = make(chan error)

func (c *Command) Main(...string) error {
	var in parse.Input

	err := redis.IsReady()
	if err != nil {
		return err
	}

	// never want to block vnet
	c.i.pub, err = publisher.New()
	if err != nil {
		return err
	}
	defer c.i.pub.Close()

	c.i.poller.pubch = make(chan string, chanDepth)
	c.i.fastPoller.pubch = make(chan string)
	go c.i.gopublish()
	// go c.i.gopublishHf()

	if c.Init != nil {
		c.init.Do(c.Init)
	}

	rpc.Register(&c.i)

	sock, err := atsock.NewRpcServer("vnetd")
	if err != nil {
		return err
	}
	defer sock.Close()

	err = redis.Assign(redis.DefaultHash+":vnet.", "vnetd", "Info")
	if err != nil {
		return err
	}

	c.i.eventPool.New = c.i.newEvent
	c.i.v.RegisterHwIfAddDelHook(c.i.hw_if_add_del)
	c.i.v.RegisterHwIfLinkUpDownHook(c.i.hw_if_link_up_down)
	c.i.v.RegisterSwIfAddDelHook(c.i.sw_if_add_del)
	c.i.v.RegisterSwIfAdminUpDownHook(c.i.sw_if_admin_up_down)
	if err = Hook(c.i.init, &c.i.v); err != nil {
		return err
	}

	in.SetString("cli { listen { no-prompt socket @vnet } }")

	signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)

	err = c.i.v.Run(&in)
	CloseHook(&c.i, &c.i.v)
	closeDone <- err
	return nil
}

func (c *Command) Close() (err error) {
	if c.i.poller.pubch != nil {
		close(c.i.poller.pubch)
	}
	if c.i.fastPoller.pubch != nil {
		close(c.i.fastPoller.pubch)
	}
	c.i.v.Quit()
	err = <-closeDone
	return
}

func (i *Info) init() {
	const (
		defaultPollInterval = 5
		defaultFastPollIntervalMilliSec = 200
	)
	i.poller.i = i
	i.fastPoller.i = i
	i.poller.addEvent(0)
	i.fastPoller.pollInterval = defaultFastPollIntervalMilliSec
	i.fastPoller.addEvent(0)
	i.poller.pollInterval = defaultPollInterval
	i.fastPoller.hostname,_ = os.Hostname()
	i.pubHwIfConfig()
	i.set("ready", "true", true)
	i.poller.pubch <- fmt.Sprint("poll.max-channel-depth: ", chanDepth)
	i.poller.pubch <- fmt.Sprint("pollInterval: ", defaultPollInterval)
	i.poller.pubch <- fmt.Sprint("pollInterval.msec: ", defaultFastPollIntervalMilliSec)
	i.poller.pubch <- fmt.Sprint("kafka-broker: ","")
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
		i.poller.pubch <- fmt.Sprint(si.Name(v), ".admin: ", parse.Enable(isUp))
	}
	return
}

func (i *Info) publish_link(hi vnet.Hi, isUp bool) {
	i.poller.pubch <- fmt.Sprint(hi.Name(&i.v), ".link: ", parse.Enable(isUp))
}

func (i *Info) hw_if_add_del(v *vnet.Vnet, hi vnet.Hi, isDel bool) (err error) {
	i.hw_if_link_up_down(v, hi, false)
	return
}

func (i *Info) hw_if_link_up_down(v *vnet.Vnet, hi vnet.Hi, isUp bool) (err error) {
	if i.hw_is_ok(hi) {
		var flag uint8 = xeth.XETH_CARRIER_OFF
		if isUp {
			flag = xeth.XETH_CARRIER_ON
		}
		// Make sure interface is known to platina-mk1 driver
		if _, found := vnet.Ports[hi.Name(v)]; found {
			index := xeth.Interface.Named(hi.Name(v)).Ifinfo.Index
			xeth.Carrier(index, flag)
		}
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
	newValue     chan string
	isReadyEvent bool
}

func (i *Info) newEvent() interface{} {
	return &event{
		i:        i,
		err:      make(chan error, 1),
		newValue: make(chan string, 1),
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
		addr   string
	)
	if e.isReadyEvent {
		e.i.poller.pubch <- fmt.Sprint(e.key, ": ", e.value)
		return
	}
	e.in.Init(nil)
	e.in.Add(e.key, e.value)
	v := &e.i.v
	switch {
	case e.in.Parse("%v.speed %v", &hi, v, &bw):
		{
			err := hi.SetSpeed(v, bw)
			h := v.HwIf(hi)
			if err == nil {
				e.newValue <- h.Speed().String()
			}
			e.err <- err
		}
	case e.in.Parse("%v.admin %v", &si, v, &enable):
		{
			err := si.SetAdminUp(v, bool(enable))
			es := "false"
			if bool(enable) {
				es = "true"
			}
			if err == nil {
				e.newValue <- es
			}
			e.err <- err
		}
	case e.in.Parse("%v.media %s", &hi, v, &media):
		{
			err := hi.SetMedia(v, media)
			h := v.HwIf(hi)
			if err == nil {
				e.newValue <- h.Media()
			}
			e.err <- err
		}
	case e.in.Parse("%v.fec %v", &hi, v, &fec):
		{
			err := ethernet.SetInterfaceErrorCorrection(v, hi, fec)
			if err == nil {
				if h, ok := v.HwIfer(hi).(ethernet.HwInterfacer); ok {
					e.newValue <- h.GetInterface().ErrorCorrectionType.String()
				} else {
					err = fmt.Errorf("error setting fec")
				}
			}
			e.err <- err
		}
	case e.in.Parse("pollInterval %f", &itv):
		if itv < 1 {
			e.err <- fmt.Errorf("pollInterval must be 1 second or longer")
		} else {
			e.i.poller.pollInterval = itv
			e.newValue <- fmt.Sprintf("%f", itv)
			e.err <- nil
		}
	case e.in.Parse("pollInterval.msec %f", &itv):
		if itv < 1 {
			e.err <- fmt.Errorf("pollInterval.msec must be 1 millisecond or longer")
		} else {
			e.i.fastPoller.pollInterval = itv
			e.newValue <- fmt.Sprintf("%f", itv)
			e.err <- nil
		}
	case e.in.Parse("kafka-broker %s", &addr):
		//e.i.initProducer(addr)
		e.newValue <- fmt.Sprintf("%s", addr)
		e.err <- nil
	default:
		e.err <- fmt.Errorf("can't set %s to %v", e.key, e.value)
	}
	e.i.eventPool.Put(e)
}

// func (i *Info) initProducer(broker string){
// 	var err error
// 	if i.producer != nil {
// 		i.producer.Close()
// 	}
// 	i.producer,err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
// 	if err != nil {
// 		fmt.Errorf("error while creating producer: %v",err)
// 	}else {
// 		go func() {
// 			for e := range i.producer.Events() {
// 				switch ev := e.(type) {
// 				case *kafka.Message:
// 					m := ev
// 					if m.TopicPartition.Error != nil {
// 						fmt.Errorf("Delivery of msg to topic %s [%d] at offset %v failed: %v \n",
// 							*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset,m.TopicPartition.Error)
// 					}
// 				default:
// 					fmt.Printf("Ignored event: %s\n", ev)
// 				}
// 			}
// 		}()
// 	}
// }
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
		newValue := <-e.newValue
		i.poller.pubch <- fmt.Sprint(e.key, ": ", newValue)
	}
	return
}

func (i *Info) gopublish() {
	for s := range i.poller.pubch {
		i.pub.Print("vnet.", s)
	}
}
//func (i *Info) gopublishHf() {
//	topic := "hf-counters"
//	for s := range i.fastPoller.pubch {
//		i.producer.ProduceChannel() <- &kafka.Message{
//			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
//			Value:          []byte(s),
//		}
//		i.fastPoller.msgCount++
//	}
//}
type hwIfConfig struct {
	speed string
	media string
	fec   string
}

var prevHwIfConfig map[string]*hwIfConfig

func (i *Info) pubHwIfConfig() {
	v := &i.v
	if prevHwIfConfig == nil {
		prevHwIfConfig = make(map[string]*hwIfConfig)
	}
	v.ForeachHwIf(UnixInterfacesOnly, func(hi vnet.Hi) {
		h := v.HwIf(hi)
		ifname := hi.Name(v)
		speed := h.Speed().String()
		media := h.Media()
		entry, found := prevHwIfConfig[ifname]
		if !found {
			entry = new(hwIfConfig)
			prevHwIfConfig[ifname] = entry
		}
		if speed != prevHwIfConfig[ifname].speed {
			prevHwIfConfig[ifname].speed = speed
			i.poller.pubch <- fmt.Sprint(ifname, ".speed: ", speed)
		}
		if media != prevHwIfConfig[ifname].media {
			prevHwIfConfig[ifname].media = media
			i.poller.pubch <- fmt.Sprint(ifname, ".media: ", media)
		}
		if h, ok := v.HwIfer(hi).(ethernet.HwInterfacer); ok {
			fec := h.GetInterface().ErrorCorrectionType.String()
			if fec != prevHwIfConfig[ifname].fec {
				prevHwIfConfig[ifname].fec = fec
				i.poller.pubch <- fmt.Sprint(ifname, ".fec: ", fec)
			}
		}
	})
}

// One per each hw/sw interface from vnet.
type ifStatsPollerInterface struct {
	lastValues map[string]uint64
	hfLastValues map[string]uint64
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
func (i *ifStatsPollerInterface) updateHf(counter string, value uint64) (delta uint64,updated bool) {
	if i.hfLastValues == nil {
		i.hfLastValues = make(map[string]uint64)
	}
	portregex := regexp.MustCompile(`(packets|bytes) *`)
	if v, ok := i.hfLastValues[counter]; ok {
		if updated = v != value; updated {
			i.hfLastValues[counter] = value
			if portregex.MatchString(counter){
				if value > v {
					delta = value - v
				}
			}else {
				delta = value
			}
		}
	} else {
		updated = true
		i.hfLastValues[counter] = value
	}
	return
}
//go:generate gentemplate -d Package=vnetd -id ifStatsPollerInterface -d VecType=ifStatsPollerInterfaceVec -d Type=ifStatsPollerInterface github.com/platinasystems/elib/vec.tmpl

type ifStatsPoller struct {
	vnet.Event
	i            *Info
	sequence     uint
	hwInterfaces ifStatsPollerInterfaceVec
	swInterfaces ifStatsPollerInterfaceVec
	pollInterval float64 // pollInterval in seconds
	pubch        chan string
}

func (p *ifStatsPoller) publish(name, counter string, value uint64) {
	p.pubch <- fmt.Sprintf("%s.%s: %d", name, counter, value)
}
func (p *ifStatsPoller) addEvent(dt float64) { p.i.v.SignalEventAfter(p, dt) }
func (p *ifStatsPoller) String() string {
	return fmt.Sprintf("redis stats poller sequence %d", p.sequence)
}
func (p *ifStatsPoller) EventAction() {
	// Schedule next event in 5 seconds; do before fetching counters so that time interval is accurate.
	p.addEvent(p.pollInterval)

	start := time.Now()
	p.pubch <- fmt.Sprint("poll.start.time: ", start.Format(time.StampMilli))
	p.pubch <- fmt.Sprint("poll.start.channel-length: ", len(p.pubch))

	p.i.pubHwIfConfig()

	// Publish all sw/hw interface counters even with zero values for first poll.
	// This was all possible counters have valid values in redis.
	// Otherwise only publish to redis when counter values change.
	includeZeroCounters := p.sequence == 0

	pubcount := func(ifname, counter string, value uint64) {
		counter = Counter(counter)
		entry := xeth.Interface.Named(ifname)
		if value != 0 && entry != nil &&
			entry.DevType == xeth.XETH_DEVTYPE_XETH_PORT {
			if _, found := vnet.Ports[ifname]; found {
				xethif := xeth.Interface.Named(ifname)
				ifindex := xethif.Ifinfo.Index
				xeth.SetStat(ifindex, counter, value)
			}

		}
		p.publish(ifname, counter, value)
	}
	p.i.v.ForeachHwIfCounter(includeZeroCounters, UnixInterfacesOnly,
		func(hi vnet.Hi, counter string, value uint64) {
			p.hwInterfaces.Validate(uint(hi))
			if p.hwInterfaces[hi].update(counter, value) && true {
				pubcount(hi.Name(&p.i.v), counter, value)
			}
		})

	p.i.v.ForeachSwIfCounter(includeZeroCounters,
		func(si vnet.Si, counter string, value uint64) {
			p.swInterfaces.Validate(uint(si))
			if p.swInterfaces[si].update(counter, value) && true {
				pubcount(si.Name(&p.i.v), counter, value)
			}
		})

	stop := time.Now()
	p.pubch <- fmt.Sprint("poll.stop.time: ", stop.Format(time.StampMilli))
	p.pubch <- fmt.Sprint("poll.stop.channel-length: ", len(p.pubch))

	p.i.v.ForeachHwIf(false, func(hi vnet.Hi) {
		h := p.i.v.HwIfer(hi)
		hw := p.i.v.HwIf(hi)
		// FIXME how to filter these in a better way?
		if strings.Contains(hw.Name(), "fe1-") ||
			strings.Contains(hw.Name(), "pg") ||
			strings.Contains(hw.Name(), "meth") {
			return
		}

		if hw.IsLinkUp() {
			sp := h.GetHwInterfaceFinalSpeed()
			// Send speed message to driver so ethtool can see it
			xethif := xeth.Interface.Named(hw.Name())
			ifindex := xethif.Ifinfo.Index
			xeth.Speed(int(ifindex), uint64(sp/1e6))
			if false {
				fmt.Println("FinalSpeed:", hw.Name(), ifindex, sp, uint64(sp/1e6))
			}
		}
	})

	p.sequence++
}

type fastIfStatsPoller struct {
	vnet.Event
	i            *Info
	sequence     uint
	hwInterfaces ifStatsPollerInterfaceVec
	swInterfaces ifStatsPollerInterfaceVec
	pollInterval float64 // pollInterval in milliseconds
	pubch        chan string
	msgCount     uint64
	hostname     string
}

func (p *fastIfStatsPoller) publish(data map[string]string) {
	for k,v := range data {
		p.pubch <- fmt.Sprintf("%s,%d,%s,%s",p.hostname,time.Now().UnixNano()/1000000,k,v)
	}
}
func (p *fastIfStatsPoller) addEvent(dt float64) { p.i.v.SignalEventAfter(p, dt) }
func (p *fastIfStatsPoller) String() string {
	return fmt.Sprintf("redis stats poller sequence %d", p.sequence)
}
func (p *fastIfStatsPoller) EventAction() {
	// Schedule next event in 200 milliseconds; do before fetching counters so that time interval is accurate.
	//p.addEvent(p.pollInterval / 1000)

	// Publish all sw/hw interface counters even with zero values for first poll.
	// This was all possible counters have valid values in redis.
	// Otherwise only publish to redis when counter values change.
	var c = make(map[string]string)
	p.i.v.ForeachHighFreqHwIfCounter(true, UnixInterfacesOnly,
		func(hi vnet.Hi, counter string, value uint64) {
			ifname := hi.Name(&p.i.v)
			p.hwInterfaces.Validate(uint(hi))
			delta, _ := p.hwInterfaces[hi].updateHf(counter, value)
			c[ifname] = c[ifname] + fmt.Sprint(delta) + ","
		})
	//if p.i.producer != nil{
	//	p.publish(c)
	//}
	p.sequence++
}
