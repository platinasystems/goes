// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ixge

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/vnet"

	"sync"
)

type addr [2]reg

func (a *addr) set(d *dev, v uint64) {
	a[0].set(d, reg(v))
	a[1].set(d, reg(v>>32))
}

type dma_regs struct {
	// [31:7] 128 byte aligned.
	descriptor_address addr

	n_descriptor_bytes reg

	// [0] rx descriptor fetch tph enable
	// [1] rx descriptor write back tph enable
	// [2] rx header data tph enable
	// [3] rx payload data tph enable
	// [5] rx/tx descriptor dca enable
	// [6] rx packet head dca enable
	// [7] rx packet tail dca enable
	// [9] rx/tx descriptor relaxed order
	// [11] rx/tx descriptor write back relaxed order
	// [13] rx/tx data write/read relaxed order
	// [15] rx head data write relaxed order
	// [31:24] apic id for cpu's cache.
	dca_control reg

	head_index reg

	// [4:0] tail buffer size (in 1k byte units)
	// [13:8] head buffer size (in 64 byte units)
	// [24:22] lo free descriptors interrupt threshold (units of 64 descriptors)
	//         interrupt is generated each time number of free descriptors is decreased to X * 64
	// [27:25] descriptor type 0 = legacy, 1 = advanced one buffer (e.g. tail),
	//   2 = advanced header splitting (head + tail), 5 = advanced header splitting (head only).
	// [28] drop if no descriptors available.
	rx_split_control reg

	tail_index reg

	// [0] rx/tx packet count
	// [1]/[2] rx/tx byte count lo/hi
	vf_stats [3]reg

	// [7:0] rx/tx prefetch threshold
	// [15:8] rx/tx host threshold
	// [24:16] rx/tx write back threshold
	// [25] rx/tx enable
	// [26] tx descriptor writeback flush
	// [30] rx strip vlan enable
	control reg

	rx_coallesce_control reg
}

type rx_dma_regs struct {
	dma_regs

	// Offset 0x30.  Only defined for queues 0-15.
	// [0] rx packet count
	// [1]/[2] rx byte count lo/hi
	// For VF, stats[1] is rx multicast packets.
	stats [3]reg

	_ reg
}

func (q *rx_dma_queue) get_regs() *rx_dma_regs {
	if q.index < 64 {
		return &q.d.regs.rx_dma0[q.index]
	} else {
		return &q.d.regs.rx_dma1[q.index-64]
	}
}

type tx_dma_regs struct {
	dma_regs

	// Offset 0x30
	_ [2]reg

	// [0] enables head write back.
	head_index_write_back_address addr
}

func (q *tx_dma_queue) get_regs() *tx_dma_regs {
	return &q.d.regs.tx_dma[q.index]
}

type dma_queue struct {
	d *dev

	mu sync.Mutex

	// Queue index.
	index uint

	// Software head/tail pointers into descriptor ring.
	// Head == tail means that ring is empty.  So we have to be careful to not fill the ring.
	len, head_index, tail_index reg
}

const n_ethernet_type_filter = 8

type dma_dev struct {
	dma_config
	rx_dev
	tx_dev
	queues_for_interrupt [vnet.NRxTx]elib.BitmapVec
}

type dma_config struct {
	rx_ring_len     uint
	rx_buffer_bytes uint
	tx_ring_len     uint
}

func (q *dma_queue) start(d *dev, dr *dma_regs) {
	// enable
	dr.control.or(d, 1<<25)

	// wait for hardware to initialize.
	for dr.control.get(d)&(1<<25) == 0 {
	}

	// Set head/tail.
	dr.head_index.set(d, q.head_index)
	dr.tail_index.set(d, q.tail_index)
}
