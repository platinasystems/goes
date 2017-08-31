// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ixge

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/vnet"

	"fmt"
	"sync/atomic"
)

func (d *dev) set_queue_interrupt_mapping(rt vnet.RxTx, queue uint, irq interrupt) {
	i0, i1 := queue/2, queue%2
	v := d.regs.interrupt.queue_mapping[i0].get(d)
	shift := 16 * i1
	if rt == vnet.Tx {
		shift += 8
	}
	m := reg(0xff) << shift
	const valid = 1 << 7
	x := (valid | (reg(irq) & 0x1f)) << shift
	v = (v &^ m) | x
	d.regs.interrupt.queue_mapping[i0].set(d, v)
	d.queues_for_interrupt[rt].Validate(uint(irq))
	b := d.queues_for_interrupt[rt][irq]
	b = b.Set(queue)
	d.queues_for_interrupt[rt][irq] = b
	queueInterruptNames[irq] = fmt.Sprintf("%v %d", rt.String(), queue)
}

func (d *dev) foreach_queue_for_interrupt(rt vnet.RxTx, i interrupt, f func(queue uint)) {
	if i < interrupt(len(d.queues_for_interrupt[rt])) {
		d.queues_for_interrupt[rt][i].ForeachSetBit(f)
	}
	return
}

type interrupt uint

const (
	irq_n_queue           = 16
	irq_queue_0           = iota
	irq_flow_director     = 16
	irq_rx_missed_packet  = 17
	irq_pcie_exception    = 18
	irq_mailbox           = 19
	irq_link_state_change = 20
	irq_link_security     = 21
	irq_manageability     = 22
	irq_time_sync         = 24
	irq_gpio_0            = 25
	irq_gpio_1            = 26
	irq_gpio_2            = 27
	irq_ecc_error         = 28
	irq_phy               = 29
	irq_tcp_timer_expired = 30
	irq_other             = 31
)

var irqStrings = [...]string{
	irq_flow_director:     "flow director",
	irq_rx_missed_packet:  "rx missed packet",
	irq_pcie_exception:    "pcie exception",
	irq_mailbox:           "mailbox",
	irq_link_state_change: "link state change",
	irq_link_security:     "link security",
	irq_manageability:     "manageability",
	irq_time_sync:         "time sync",
	irq_gpio_0:            "gpio 0",
	irq_gpio_1:            "gpio 1",
	irq_gpio_2:            "gpio 2",
	irq_ecc_error:         "ecc error",
	irq_phy:               "phy",
	irq_tcp_timer_expired: "tcp timer expired",
	irq_other:             "other",
}

var queueInterruptNames = [16]string{}

func (i interrupt) String() (s string) {
	if i < irq_n_queue {
		if queueInterruptNames[i] != "" {
			s = fmt.Sprintf("queue %s", queueInterruptNames[i])
		} else {
			s = fmt.Sprintf("queue %d", i)
		}
	} else {
		s = elib.StringerHex(irqStrings[:], int(i))
	}
	return
}

func (d *dev) get_link_state() bool { return d.regs.xge_mac.link_status.get(d)&(1<<30) != 0 }

// Signal link state change to vnet event handler.
func (d *dev) link_state_change() {
	d.SignalEvent(&vnet.LinkStateEvent{
		Hi:   d.HwIf.Hi(),
		IsUp: d.get_link_state(),
	})
}

type irq_event struct {
	name elog.StringRef
	irq  interrupt
}

func (e *irq_event) Elog(l *elog.Log) { l.Logf("%s irq %s", e.name, e.irq) }

func (d *dev) interrupt_dispatch(i uint) {
	irq := interrupt(i)
	if elog.Enabled() {
		e := irq_event{name: d.elog_name, irq: irq}
		elog.Add(&e)
	}
	switch {
	case irq < irq_n_queue:
		d.foreach_queue_for_interrupt(vnet.Rx, irq, d.rx_queue_interrupt)
		d.foreach_queue_for_interrupt(vnet.Tx, irq, d.tx_queue_interrupt)
	case irq == irq_link_state_change:
		d.link_state_change()
	}
}

// Atomically get and set interrupt status.
func (d *dev) get_irq_status() (s uint32) {
	for {
		s = atomic.LoadUint32(&d.irq_status)
		if atomic.CompareAndSwapUint32(&d.irq_status, s, 0) {
			break
		}
	}
	return
}

func (d *dev) set_irq_status(s uint32) {
	for {
		old := atomic.LoadUint32(&d.irq_status)
		new := old | s
		if atomic.CompareAndSwapUint32(&d.irq_status, old, new) {
			break
		}
	}
}

func (d *dev) InterfaceInput(out *vnet.RefOut) {
	if !d.interruptsEnabled {
		for qi := range d.tx_queues {
			d.tx_queue_interrupt(uint(qi))
		}
		for qi := range d.rx_queues {
			d.rx_queue_interrupt(uint(qi))
		}
	} else {
		d.is_active = uint(0)
		s := d.get_irq_status()
		d.out = out
		elib.Word(s).ForeachSetBit(d.interrupt_dispatch)

		inc := int32(0)
		if d.is_active == 0 {
			inc--
		}
		d.active_lock.Lock()
		x := atomic.AddInt32(&d.active_count, -1)
		d.Activate(x > 0)
		d.active_lock.Unlock()

		if elog.Enabled() {
			e := irq_elog{
				name:   d.elog_name,
				status: s,
				active: x,
			}
			elog.Add(&e)
		}
	}
}

func (d *dev) Interrupt() {
	if d.interruptsEnabled {
		s := d.regs.interrupt.status_write_1_to_set.get(d)
		d.regs.interrupt.status_write_1_to_clear.set(d, s)
		d.set_irq_status(uint32(s))
		d.active_lock.Lock()
		x := atomic.AddInt32(&d.active_count, 1)
		d.Activate(true)
		d.active_lock.Unlock()

		if elog.Enabled() {
			e := irq_elog{
				name:   d.elog_name,
				status: uint32(s),
				active: x,
				is_irq: true,
			}
			elog.Add(&e)
		}
	}
}

type irq_elog struct {
	name   elog.StringRef
	status uint32
	active int32
	is_irq bool
}

func (e *irq_elog) Elog(l *elog.Log) {
	what := "input"
	if e.is_irq {
		what = "interrupt"
	}
	l.Logf("%s %s status 0x%x active %d", e.name, what, e.status, e.active)
	elib.Word(e.status).ForeachSetBit(func(i uint) {
		l.Logf("%s irq %s", e.name, interrupt(i))
	})
}

func (d *dev) InterruptEnable(enable bool) {
	all := ^reg(0) &^ (1 << 31) // all except "other"
	if enable {
		d.regs.interrupt.enable_write_1_to_set.set(d, all)
	} else {
		d.regs.interrupt.enable_write_1_to_clear.set(d, all)
	}
	d.interruptsEnabled = enable
}
