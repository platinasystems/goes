// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ixge

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/vnet"

	"fmt"
	"unsafe"
)

type tx_dma_queue struct {
	dma_queue
	tx_desc tx_descriptor_vec
	desc_id elib.Index

	head_index_write_back    *reg
	head_index_write_back_id elib.Index

	txRing vnet.TxDmaRing
}

//go:generate gentemplate -d Package=ixge -id tx_dma_queue -d VecType=tx_dma_queue_vec -d Type=tx_dma_queue github.com/platinasystems/go/elib/vec.tmpl

type tx_descriptor struct {
	buffer_address      uint64
	n_bytes_this_buffer uint16
	status0             uint16
	status1             uint32
}

//go:generate gentemplate -d Package=ixge -id tx_descriptor -d Type=tx_descriptor -d VecType=tx_descriptor_vec github.com/platinasystems/go/elib/hw/dma_mem.tmpl

const (
	tx_desc_status0_log2_is_end_of_packet   = 8 + 0
	tx_desc_status0_is_end_of_packet        = 1 << tx_desc_status0_log2_is_end_of_packet
	tx_desc_status0_insert_crc              = 1 << (8 + 1)
	tx_desc_status0_log2_report_status      = (8 + 3)
	tx_desc_status0_report_status           = (1 << tx_desc_status0_log2_report_status)
	tx_desc_status0_is_advanced             = 1 << (8 + 5)
	tx_desc_status0_vlan_enable             = 1 << (8 + 6)
	tx_desc_status1_is_owned_by_software    = 1 << 0
	tx_desc_status1_insert_tcp_udp_checksum = 1 << (8 + 1)
	tx_desc_status1_insert_ip4_checksum     = 1 << (8 + 0)

	// Only valid if is_advanced is set.
	tx_desc_status0_advanced_context = 2 << 4
	tx_desc_status0_advanced_data    = 3 << 4
)

func (e *tx_descriptor) String() (s string) {
	d := *e
	s0, s1 := d.status0, d.status1
	if s1&tx_desc_status1_is_owned_by_software != 0 {
		s += "sw: "
	} else {
		s += "hw: "
	}
	s += fmt.Sprintf("buffer %x, bytes %d", d.buffer_address, d.n_bytes_this_buffer)
	if s0&tx_desc_status0_is_end_of_packet != 0 {
		s += ", eop"
	}
	if s0&tx_desc_status0_report_status != 0 {
		s += ", report-status"
	}
	if s0&tx_desc_status0_is_advanced != 0 {
		s += ", advanced"
	}
	if s0&tx_desc_status0_vlan_enable != 0 {
		s += ", vlan-enable"
	}
	if s1&tx_desc_status1_insert_tcp_udp_checksum != 0 {
		s += ", insert tcp/udp checksum"
	}
	if s1&tx_desc_status1_insert_ip4_checksum != 0 {
		s += ", insert ip4 checksum"
	}
	return
}

// Descriptors must be 128 bit aligned.
const log2DescriptorAlignmentBytes = 7

func (d *dev) tx_dma_init(queue uint) {
	// Add 8 since head == tail means ring is empty and we want to handle
	// up to vnet.MaxOutstandingTxRefs descriptors.
	// Empirical tests show that 8 seems to be minimum number of descriptors to add.
	// Likely that some chips require ring to be multiple of 8 descriptors long.
	d.tx_ring_len = vnet.MaxOutstandingTxRefs + 8
	q := d.tx_queues.Validate(queue)
	q.d = d
	q.index = queue
	q.txRing.Init(d.m.Vnet)

	q.tx_desc, q.desc_id = tx_descriptorAllocAligned(d.tx_ring_len, log2DescriptorAlignmentBytes)
	for i := range q.tx_desc {
		q.tx_desc[i] = tx_descriptor{}
	}
	q.len = reg(d.tx_ring_len)

	dr := q.get_regs()
	dr.descriptor_address.set(d, uint64(q.tx_desc[0].PhysAddress()))
	n_desc := reg(len(q.tx_desc))
	dr.n_descriptor_bytes.set(d, n_desc*reg(unsafe.Sizeof(q.tx_desc[0])))

	{
		const base = tx_desc_status0_insert_crc
		d.tx_desc_status0_by_next_valid_flag[0] = base | tx_desc_status0_is_end_of_packet
		d.tx_desc_status0_by_next_valid_flag[vnet.NextValid] = base
	}

	// Allocate DMA memory for tx head index write back.
	{
		var b []byte
		b, q.head_index_write_back_id, _, _ = hw.DmaAlloc(4)
		p := unsafe.Pointer(&b[0])
		q.head_index_write_back = (*reg)(p)
		// Must initialize so interrupt routine will not read stale data.
		*q.head_index_write_back = 0
		const valid = 1
		dr.head_index_write_back_address.set(d, valid|uint64(hw.DmaPhysAddress(uintptr(p))))
	}

	hw.MemoryBarrier()

	{
		v := dr.control.get(d)
		// prefetch threshold
		v = (v &^ (0xff << 0)) | ((64 - 4) << 0)
		// host threshold
		v = (v &^ (0xff << 8)) | (4 << 8)
		// writeback theshold
		v = (v &^ (0xff << 16)) | (0 << 16)
		dr.control.set(d, v)
	}

	{
		v := dr.cache_control.get(d)
		v |= dma_cache_control_relaxed_ordering_rx_tx_desc_fetch |
			dma_cache_control_relaxed_ordering_rx_tx_desc_writeback |
			dma_cache_control_relaxed_ordering_rx_tx_data
		if d.have_tph {
			v |= dma_cache_control_tph_rx_tx_desc_fetch |
				dma_cache_control_tph_rx_tx_desc_writeback |
				dma_cache_control_tph_rx_tail_data_tx_data
		}
		dr.cache_control.set(d, v)
	}
}

func (d *dev) tx_dma_enable(queue uint, enable bool) {
	q := &d.tx_queues[queue]
	dr := q.get_regs()
	if enable {
		d.regs.tx_dma_control.or(d, 1<<0)
		q.start(d, &dr.dma_regs)
	} else {
		panic("not yet")
	}
}

func (d *dev) set_tx_descriptor(rs []vnet.Ref, ds []tx_descriptor, ri, di reg) {
	r0, d0 := &rs[ri+0], &ds[di+0]
	d0.buffer_address = uint64(r0.DataPhys())
	d0.n_bytes_this_buffer = uint16(r0.DataLen())
	f0 := r0.NextValidFlag()
	d0.status0 = d.tx_desc_status0_by_next_valid_flag[f0]
	// Owned by hardware.
	d0.status1 = 0
}

func (d *dev) set_4_tx_descriptors(rs []vnet.Ref, ds []tx_descriptor, ri, di reg) {
	r0, r1, r2, r3 := &rs[ri+0], &rs[ri+1], &rs[ri+2], &rs[ri+3]
	d0, d1, d2, d3 := &ds[di+0], &ds[di+1], &ds[di+2], &ds[di+3]

	d0.buffer_address = uint64(r0.DataPhys())
	d1.buffer_address = uint64(r1.DataPhys())
	d2.buffer_address = uint64(r2.DataPhys())
	d3.buffer_address = uint64(r3.DataPhys())

	d0.n_bytes_this_buffer = uint16(r0.DataLen())
	d1.n_bytes_this_buffer = uint16(r1.DataLen())
	d2.n_bytes_this_buffer = uint16(r2.DataLen())
	d3.n_bytes_this_buffer = uint16(r3.DataLen())

	f0, f1, f2, f3 := r0.NextValidFlag(), r1.NextValidFlag(), r2.NextValidFlag(), r3.NextValidFlag()

	d0.status0 = d.tx_desc_status0_by_next_valid_flag[f0]
	d1.status0 = d.tx_desc_status0_by_next_valid_flag[f1]
	d2.status0 = d.tx_desc_status0_by_next_valid_flag[f2]
	d3.status0 = d.tx_desc_status0_by_next_valid_flag[f3]

	d0.status1 = 0
	d1.status1 = 0
	d2.status1 = 0
	d3.status1 = 0
}

func (d *dev) set_tx_descriptors(rs []vnet.Ref, ds []tx_descriptor, ri0, di0, rn, dn reg) (ri, di, nd reg) {
	ri, di = ri0, di0
	for ri+4 <= rn && di+4 <= dn {
		d.set_4_tx_descriptors(rs, ds, ri, di)
		ri += 4
		di += 4
	}
	for ri < rn && di < dn {
		d.set_tx_descriptor(rs, ds, ri, di)
		ri += 1
		di += 1
	}
	nd = ri - ri0
	return
}

type tx_dev struct {
	tx_queues                          tx_dma_queue_vec
	tx_desc_status0_by_next_valid_flag [vnet.NextValid + 1]uint16
}

func (d *dev) InterfaceOutput(i *vnet.TxRefVecIn) { d.tx_queues[0].output(i) }

func (q *tx_dma_queue) output(in *vnet.TxRefVecIn) {
	d := q.d
	nr := reg(in.Len())

	head, tail := q.head_index, q.tail_index

	// Free slots are after tail and before head.
	n_free := head - tail
	if int32(n_free) <= 0 {
		n_free += q.len
	}
	// Head == tail means empty ring so we can only fill LEN - 1 descriptors in ring.
	if n_free == q.len {
		n_free--
	}

	// No room?
	if n_free < nr {
		// Should never happen since MaxOutstandingTxRefs should be enforced.
		panic(fmt.Errorf("%s: tx ring full", d.Name()))
	}

	ds, rs := q.tx_desc, in.Refs

	ri, n_tx := reg(0), reg(0)

	// From tail to end of ring.
	di := tail
	n_end := n_free
	if tail+n_end > q.len {
		n_end = q.len - tail
	}
	if n_end > 0 {
		var nd reg
		ri, di, nd = d.set_tx_descriptors(rs, ds, ri, di, nr, di+n_end)
		n_free -= nd
		n_tx += nd
	}

	// From start of ring to head.
	n_start := n_free
	if n_start > head {
		n_start = head
	}
	if n_start > 0 && ri < nr {
		var nd reg
		ri, di, nd = d.set_tx_descriptors(rs, ds, ri, 0, nr, n_start)
		n_free -= nd
		n_tx += nd
	}

	// Ring wrap.
	if di >= q.len {
		di = 0
	}

	if elog.Enabled() {
		e := tx_output_elog{
			name:     d.elog_name,
			n_done:   n_tx,
			head:     head,
			old_tail: tail,
			new_tail: di,
		}
		e.count = di - head
		if int32(e.count) < 0 {
			e.count += q.len
		}
		elog.Add(&e)
	}

	// Re-start dma engine when tail advances.
	if di != q.tail_index {
		q.tail_index = di

		// Report status when done with this vector.
		// This triggers head index write back.
		i := di - 1
		if di == 0 {
			i = q.len - 1
		}
		ds[i].status0 |= tx_desc_status0_report_status

		hw.MemoryBarrier()

		q.txRing.ToInterrupt <- in

		dr := q.get_regs()
		dr.tail_index.set(d, di)
	}
}

type tx_output_elog struct {
	name               elog.StringRef
	head               reg
	old_tail, new_tail reg
	n_done             reg
	count              reg
}

func (e *tx_output_elog) Elog(l *elog.Log) {
	l.Logf("%s tx %d tail %d -> %d head %d, count %d", e.name, e.n_done, e.old_tail, e.new_tail, e.head, e.count)
}

func (d *dev) tx_queue_interrupt(queue uint) {
	q := &d.tx_queues[0]

	// Mutually excludes real interrupt and polled calls from interfering with each other.
	q.mu.Lock()
	defer q.mu.Unlock()

	di := *q.head_index_write_back
	n_advance := di - q.head_index
	if int32(n_advance) < 0 {
		n_advance += q.len
	}
	q.head_index = di
	if elog.Enabled() {
		dr := q.get_regs()
		tail := dr.tail_index.get(d)
		e := tx_advance_elog{
			name:      d.elog_name,
			n_advance: n_advance,
			head:      di,
			tail:      tail,
		}
		e.count = tail - di
		if int32(e.count) < 0 {
			e.count += q.len
		}
		elog.Add(&e)
	}

	// Remain active until tx ring is empty and we didn't advance.
	if !(n_advance == 0 && q.head_index == q.tail_index) {
		d.is_active += 1
	}

	q.txRing.InterruptAdvance(uint(n_advance))
}

type tx_advance_elog struct {
	name       elog.StringRef
	head, tail reg
	count      reg
	n_advance  reg
}

func (e *tx_advance_elog) Elog(l *elog.Log) {
	l.Logf("%s tx advance %d head %d tail %d count %d", e.name, e.n_advance, e.head, e.tail, e.count)
}
