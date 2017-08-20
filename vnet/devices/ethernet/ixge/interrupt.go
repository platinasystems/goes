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

func (i interrupt) String() (s string) {
	if i < irq_n_queue {
		s = fmt.Sprintf("queue %d", i)
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

func (d *dev) interrupt_dispatch(i uint) {
	irq := interrupt(i)
	elog.F("%s irq %d", d.elog_name, i)
	switch {
	case irq < irq_n_queue:
		d.foreach_queue_for_interrupt(vnet.Rx, irq, d.rx_queue_interrupt)
		d.foreach_queue_for_interrupt(vnet.Tx, irq, d.tx_queue_interrupt)
	case irq == irq_link_state_change:
		d.link_state_change()
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
		// Get status and ack interrupt.
		d.irq_sequence++
		s := d.regs.interrupt.status_write_1_to_set.get(d)
		if s != 0 {
			d.regs.interrupt.status_write_1_to_clear.set(d, s)
			d.out = out
			elib.Word(s).ForeachSetBit(d.interrupt_dispatch)
		}

		// Poll any queues that need polling and have not been polled this interrupt.
		for qi := range d.rx_queues {
			q := &d.rx_queues[qi]
			if q.needs_polling && q.irq_sequence != d.irq_sequence {
				d.rx_queue_interrupt(uint(qi))
			}
		}
		for qi := range d.tx_queues {
			q := &d.tx_queues[qi]
			if q.needs_polling && q.irq_sequence != d.irq_sequence {
				d.tx_queue_interrupt(uint(qi))
			}
		}

		d.Activate(atomic.AddInt32(&d.active_count, -1) > 0)
	}
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

func (d *dev) Interrupt() {
	if d.interruptsEnabled {
		d.Activate(true)
		atomic.AddInt32(&d.active_count, 1)
	}
}
