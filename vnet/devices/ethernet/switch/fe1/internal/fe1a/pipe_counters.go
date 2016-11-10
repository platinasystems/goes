// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/pipemem"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"math/rand"
	"sync"
)

type rx_tx_pipe_reg32 m.Reg32

func (r *rx_tx_pipe_reg32) geta(q *DmaRequest, b sbus.Block, c sbus.AccessType, v *uint32) {
	(*m.Reg32)(r).Get(&q.DmaRequest, 0, b, c, v)
}
func (r *rx_tx_pipe_reg32) seta(q *DmaRequest, b sbus.Block, c sbus.AccessType, v uint32) {
	(*m.Reg32)(r).Set(&q.DmaRequest, 0, b, c, v)
}
func (r *rx_tx_pipe_reg32) get(q *DmaRequest, b sbus.Block, v *uint32) {
	r.geta(q, b, sbus.Duplicate, v)
}
func (r *rx_tx_pipe_reg32) set(q *DmaRequest, b sbus.Block, v uint32) { r.seta(q, b, sbus.Duplicate, v) }
func (r *rx_tx_pipe_reg32) getDo(q *DmaRequest, b sbus.Block, c sbus.AccessType) (v uint32) {
	r.geta(q, b, c, &v)
	q.Do()
	return
}

type rx_tx_pipe_reg64 m.Reg64

func (r *rx_tx_pipe_reg64) geta(q *DmaRequest, b sbus.Block, c sbus.AccessType, v *uint64) {
	(*m.Reg64)(r).Get(&q.DmaRequest, 0, b, c, v)
}
func (r *rx_tx_pipe_reg64) seta(q *DmaRequest, b sbus.Block, c sbus.AccessType, v uint64) {
	(*m.Reg64)(r).Set(&q.DmaRequest, 0, b, c, v)
}
func (r *rx_tx_pipe_reg64) get(q *DmaRequest, b sbus.Block, v *uint64) {
	r.geta(q, b, sbus.Duplicate, v)
}
func (r *rx_tx_pipe_reg64) set(q *DmaRequest, b sbus.Block, v uint64) { r.seta(q, b, sbus.Duplicate, v) }
func (r *rx_tx_pipe_reg64) getDo(q *DmaRequest, b sbus.Block, c sbus.AccessType) (v uint64) {
	r.geta(q, b, c, &v)
	q.Do()
	return
}

const (
	pipe_counter_control_enable        = 1 << 1
	pipe_counter_control_clear_on_read = 1 << 2
)

type pipe_counter_4pool_control struct {
	control [4]rx_tx_pipe_reg32

	packet_attribute_selector_key [4]rx_tx_pipe_reg64

	eviction_control   [4]eviction_control_reg
	eviction_lfsr_seed [4]eviction_seed_reg
	eviction_threshold [4]eviction_threshold_reg

	_ [0x80 - 0x14]rx_tx_pipe_reg32

	offset_table_control [4]rx_tx_pipe_reg32

	_ [0x100 - 0x84]rx_tx_pipe_reg32
}

type pipe_counter_1pool_control struct {
	control            rx_tx_pipe_reg32
	eviction_control   eviction_control_reg
	eviction_threshold eviction_threshold_reg
	eviction_seed      eviction_seed_reg
}

type pipe_counter_config struct {
	q                      *DmaRequest
	is_clear_on_read       bool
	eviction_threshold     vnet.CombinedCounter
	random_eviction_enable bool
}

func (r *pipe_counter_1pool_control) enable(t *fe1a, c pipe_counter_config) {
}

type eviction_control_reg rx_tx_pipe_reg32

func (r *eviction_control_reg) set(c pipe_counter_config, b sbus.Block, pipe uint, memory_id int) {
	mode := 2
	if c.random_eviction_enable {
		mode = 1
	}
	v := uint32(mode << 8)
	v |= uint32(memory_id) | uint32(pipe)<<6
	(*rx_tx_pipe_reg32)(r).seta(c.q, b, sbus.Unique(pipe), v)
}

type eviction_threshold_reg rx_tx_pipe_reg64

func (r *eviction_threshold_reg) set(c pipe_counter_config, b sbus.Block, pipe uint) {
	v := c.eviction_threshold.Packets | c.eviction_threshold.Bytes<<26
	(*rx_tx_pipe_reg64)(r).seta(c.q, b, sbus.Unique(pipe), v)
}

type eviction_seed_reg rx_tx_pipe_reg64

func (r *eviction_seed_reg) set(c pipe_counter_config, b sbus.Block, pipe uint) {
	v := uint64(rand.Int())
	(*rx_tx_pipe_reg64)(r).seta(c.q, b, sbus.Unique(pipe), v)
}

func (r *pipe_counter_1pool_control) set(c pipe_counter_config, b sbus.Block, pipe uint, memory_id int) {
	r.eviction_control.set(c, b, pipe, memory_id)
	r.eviction_threshold.set(c, b, pipe)
	r.eviction_seed.set(c, b, pipe)
}

func (r *pipe_counter_4pool_control) set(c pipe_counter_config, b sbus.Block, pipe uint, memory_id, i int) {
	r.eviction_control[i].set(c, b, pipe, memory_id)
	r.eviction_threshold[i].set(c, b, pipe)
	r.eviction_lfsr_seed[i].set(c, b, pipe)
}

const n_pipe_counter_mode = 4 // 2 bits of mode

type pipe_counter_4pool_mems struct {
	pool_offset_tables [4]struct {
		entries [n_pipe_counter_mode][256]m.Mem32
		_       [0x1000 - n_pipe_counter_mode*256]m.MemElt
	}

	pool_counters [4][0x4000]pipe_counter_mem

	_ [0x20000 - 0x14000]byte
}

type rx_pipe_pipe_counter_mem m.MemElt
type tx_pipe_pipe_counter_mem m.MemElt
type pipe_counter_mem m.MemElt

type pipe_counter_entry struct{ vnet.CombinedCounter }

func (e *pipe_counter_entry) MemBits() int { return 68 }
func (e *pipe_counter_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint64(&e.Packets, b, i+25, i, isSet)
	i = m.MemGetSetUint64(&e.Bytes, b, i+33, i, isSet)
}

func (e *pipe_counter_mem) geta(q *DmaRequest, b sbus.Block, pipe uint, v *pipe_counter_entry) {
	(*m.MemElt)(e).MemDmaGet(&q.DmaRequest, v, b, sbus.Unique(pipe))
}
func (e *pipe_counter_mem) seta(q *DmaRequest, b sbus.Block, pipe uint, v *pipe_counter_entry) {
	(*m.MemElt)(e).MemDmaSet(&q.DmaRequest, v, b, sbus.Unique(pipe))
}

func (e *rx_pipe_pipe_counter_mem) geta(q *DmaRequest, rx_pipe uint, v *pipe_counter_entry) {
	(*m.MemElt)(e).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, sbus.Unique(rx_pipe))
}
func (e *rx_pipe_pipe_counter_mem) seta(q *DmaRequest, rx_pipe uint, v *pipe_counter_entry) {
	(*m.MemElt)(e).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, sbus.Unique(rx_pipe))
}
func (e *tx_pipe_pipe_counter_mem) geta(q *DmaRequest, tx_pipe uint, v *pipe_counter_entry) {
	(*m.MemElt)(e).MemDmaGet(&q.DmaRequest, v, BlockTxPipe, sbus.Unique(tx_pipe))
}
func (e *tx_pipe_pipe_counter_mem) seta(q *DmaRequest, tx_pipe uint, v *pipe_counter_entry) {
	(*m.MemElt)(e).MemDmaSet(&q.DmaRequest, v, BlockTxPipe, sbus.Unique(tx_pipe))
}

const (
	n_pipe_counter_pool_rx_pipe              = 20
	n_pipe_counter_pool_tx_pipe              = 4
	pipe_counter_memory_id_pool_rx_pipe      = 1
	pipe_counter_memory_id_pool_tx_pipe      = pipe_counter_memory_id_pool_rx_pipe + n_pipe_counter_pool_rx_pipe
	pipe_counter_memory_id_tx_pipe_perq_pool = pipe_counter_memory_id_pool_tx_pipe + n_pipe_counter_pool_tx_pipe
	pipe_counter_memory_id_tx_pipe_txf_pool  = pipe_counter_memory_id_tx_pipe_perq_pool + 1
	n_pipe_counter_memory_id                 = 26
)

func (t *fe1a) pipe_counter_init() {
	q := t.getDmaReq()
	fm := &t.pipe_counter_main

	fm.rx_pipe.pools = make([]pipe_counter_pool, n_pipe_counter_pool_rx_pipe)
	fm.tx_pipe.pools = make([]pipe_counter_pool, n_pipe_counter_pool_tx_pipe+1) // 1 extra for txf

	for i := range fm.rx_pipe.pools {
		max_len := uint(4096)
		if i >= 8 {
			max_len = 512
		}
		fm.rx_pipe.pools[i].Init(n_rx_pipe, max_len)
	}
	for i := range fm.tx_pipe.pools {
		max_len := uint(4096)
		if i >= 2 {
			max_len = 1024
		}
		fm.tx_pipe.pools[i].Init(n_tx_pipe, max_len)
	}

	// Enable central eviction fifo.
	{
		v := uint32(1 << 0) // enable bit
		v |= uint32(n_pipe_counter_memory_id << 1)
		t.rx_pipe_regs.pipe_counter_eviction_control.set(q, v)
		q.Do()
	}

	// Enable threshold based eviction.
	{
		c := pipe_counter_config{
			q: q,
			random_eviction_enable: false,
			eviction_threshold: vnet.CombinedCounter{
				Packets: (1 << 26) / 2, // 26 bits
				Bytes:   (1 << 34) / 2, // 34 bits
			},
		}

		for pipe := uint(0); pipe < n_pipe; pipe++ {
			// Rx pipe counters.
			mem_id := pipe_counter_memory_id_pool_rx_pipe
			for i := range t.rx_pipe_regs.pipe_counter {
				for j := range t.rx_pipe_regs.pipe_counter[i].eviction_control {
					t.rx_pipe_regs.pipe_counter[i].set(c, BlockRxPipe, pipe, mem_id, j)
					mem_id++
				}
			}

			// Tx pipe counters.
			t.tx_pipe_regs.perq_counter.set(c, BlockTxPipe, pipe, pipe_counter_memory_id_tx_pipe_perq_pool)
			t.tx_pipe_regs.txf_counter.set(c, BlockTxPipe, pipe, pipe_counter_memory_id_tx_pipe_txf_pool)
			for i := range t.tx_pipe_regs.pipe_counter.eviction_control {
				t.tx_pipe_regs.pipe_counter.set(c, BlockTxPipe, pipe, pipe_counter_memory_id_pool_tx_pipe+i, i)
			}
		}
		q.Do()
	}

	// Enable all counters and set to clear on read.
	{
		const v uint32 = pipe_counter_control_enable | pipe_counter_control_clear_on_read
		t.tx_pipe_regs.perq_counter.control.set(q, BlockTxPipe, v)
		t.tx_pipe_regs.txf_counter.control.set(q, BlockTxPipe, v)
		for i := range t.tx_pipe_regs.pipe_counter.control {
			t.tx_pipe_regs.pipe_counter.control[i].set(q, BlockTxPipe, v)
		}
		for i := range t.rx_pipe_regs.pipe_counter {
			for j := range t.rx_pipe_regs.pipe_counter[i].control {
				t.rx_pipe_regs.pipe_counter[i].control[j].set(q, BlockRxPipe, v)
			}
		}
		q.Do()
	}

	t.pipe_counter_fifo_dma_start()
}

func (t *fe1a) pipe_counter_init_offset_table(b sbus.Block, pool uint) {
	q := t.getDmaReq()
	p0, p1 := pool/4, pool%4
	for mode := 0; mode < 4; mode++ {
		for i := 0; i < 256; i++ {
			const enable = 1
			if b == BlockRxPipe {
				if p0 < 3 {
					t.rx_pipe_mems.pipe_counter0[p0].pool_offset_tables[p1].entries[mode][i].Set(&q.DmaRequest, b, sbus.Duplicate, enable)
				} else {
					t.rx_pipe_mems.pipe_counter1[p0-3].pool_offset_tables[p1].entries[mode][i].Set(&q.DmaRequest, b, sbus.Duplicate, enable)
				}
			} else {
				t.tx_pipe_mems.pipe_counter.pool_offset_tables[p1].entries[mode][i].Set(&q.DmaRequest, b, sbus.Duplicate, enable)
			}
		}
		q.Do()
	}
}

// Start fifo dma; must choose non-zero channel.  Zero is reserved for l2 mod fifo.
const pipe_counter_fifo_dma_channel = 1

func (t *fe1a) pipe_counter_fifo_dma_start() {
	fm := &t.pipe_counter_main

	fm.resultFifo = make(chan sbus.FifoDmaData, 64)
	go func(t *fe1a) {
		for d := range fm.resultFifo {
			t.handle_fifo_data(d)
		}
	}(t)

	t.CpuMain.FifoDmaInit(fm.resultFifo,
		pipe_counter_fifo_dma_channel, // channel
		t.rx_pipe_mems.pipe_counter_eviction_fifo[0].Address(),
		sbus.Command{
			Opcode:     sbus.FifoPop,
			Block:      BlockRxPipe,
			AccessType: sbus.Unique(0),
			Size:       4,
		},
		// Number of bits in pipe counter eviction fifo.
		84,
		// Log2 # of entries in host buffer.
		10)
}

func (t *fe1a) handle_fifo_data(d sbus.FifoDmaData) {
	if len(d.Data) == 0 {
		return
	}
	fm := &t.pipe_counter_main
	// Mutually exclude calls from interrupt and explicit calls via pipe_counter_eviction_fifo_sync.
	fm.mu.Lock()
	defer fm.mu.Unlock()
	for i := 0; i < len(d.Data); i += 3 {
		t.decode_eviction_fifo(d.Data[i : i+3])
	}
	d.Free()
}

// Poll eviction fifo for entries.
func (t *fe1a) pipe_counter_eviction_fifo_sync() {
	for {
		d := t.CpuMain.FifoDmaSync(pipe_counter_fifo_dma_channel)
		if len(d.Data) == 0 {
			break
		}
		t.handle_fifo_data(d)
	}
}

type pipe_counter struct {
	ref       pipemem.Ref
	dma_value pipe_counter_entry
	value     pipe_counter_entry
}

//go:generate gentemplate -d Package=fe1a -id pipe_counter -d VecType=pipe_counter_vec -d Type=pipe_counter github.com/platinasystems/go/elib/vec.tmpl

type pipe_counter_pool struct {
	mu sync.Mutex
	pipemem.Pool
	counters [n_pipe]pipe_counter_vec
}

type pipe_counter_pipe_main struct {
	pools []pipe_counter_pool
}

type pipe_counter_main struct {
	rx_pipe, tx_pipe pipe_counter_pipe_main
	mu               sync.Mutex
	resultFifo       chan sbus.FifoDmaData
}

func (m *pipe_counter_main) get_pipe(b sbus.Block) (pm *pipe_counter_pipe_main) {
	pm = &m.rx_pipe
	if b == BlockTxPipe {
		pm = &m.tx_pipe
	}
	return
}

// Pool usage.
const (
	// RxPipe: pools 0-7 have 4k counters; pools 8-19 have 512 counters
	pipe_counter_pool_rx_l3_interface = 0
	pipe_counter_pool_rx_port_table   = 8
	// Tx_pipe: pools 0-1 have 4k counters; pools 2-3 have 1k counters.
	pipe_counter_pool_tx_l3_interface = 0
	pipe_counter_pool_tx_adjacency    = 1
	pipe_counter_pool_tx_port_table   = 2
)

// Zero index is never valid (enforced by hardware).
func (x *pipe_counter_ref_entry) is_valid() bool { return x.index != 0 }

func (x *pipe_counter_ref_entry) alloc(t *fe1a, poolIndex, pipeMask uint, b sbus.Block) (ref pipemem.Ref, ok bool) {
	fm := &t.pipe_counter_main
	pm := fm.get_pipe(b)

	pool := &pm.pools[poolIndex]
	if pool.Len() == 0 {
		t.pipe_counter_init_offset_table(b, poolIndex)

		// Zero offsets are ignored by hardware.  So allocate them in all pipes.
		pool.Get(0xf)
	}

	if ref, ok = pool.Get(pipeMask); !ok {
		return
	}
	ri, _ := ref.Get()

	// Protect counter resize from fifo interrupt.
	pool.mu.Lock()
	for pipe := uint(0); pipe < n_pipe; pipe++ {
		pool.counters[pipe].Validate(ri)
	}
	pool.mu.Unlock()

	x.pool = uint8(poolIndex)
	x.mode = 0
	x.index = uint16(ri)
	x.ref = ref
	ok = true
	return
}

func (x *pipe_counter_ref_entry) free(t *fe1a, b sbus.Block) {
	fm := &t.pipe_counter_main
	pm := fm.get_pipe(b)
	pool := &pm.pools[x.pool]
	pool.Put(x.ref)
}

func (x *pipe_counter_ref_entry) get_value(t *fe1a, b sbus.Block) (v vnet.CombinedCounter) {
	fm := &t.pipe_counter_main
	pm := fm.get_pipe(b)
	pool := &pm.pools[x.pool]
	i, m := x.ref.Get()
	for pipe := uint(0); pipe < n_pipe; pipe++ {
		if m&(1<<pipe) != 0 {
			v.Add(&pool.counters[pipe][i].value.CombinedCounter)
		}
	}
	return
}

// NB: does not sync eviction fifo.
func (x *pipe_counter_ref_entry) update_value(t *fe1a, pipe uint, b sbus.Block) (v pipe_counter_entry) {
	fm := &t.pipe_counter_main
	pm := fm.get_pipe(b)
	pool := &pm.pools[x.pool]
	c := &pool.counters[pipe][x.index]
	c.dma_value.Zero()
	q := t.getDmaReq()
	if b == BlockTxPipe {
		if x.pool < n_pipe_counter_pool_tx_pipe {
			t.tx_pipe_mems.pipe_counter.pool_counters[x.pool][x.index].geta(q, BlockTxPipe, pipe, &c.dma_value)
		} else {
			t.tx_pipe_mems.txf_counter_table[x.index].geta(q, pipe, &c.dma_value)
		}
	} else {
		p0, p1 := x.pool/4, x.pool%4
		if p0 < 3 {
			t.rx_pipe_mems.pipe_counter0[p0].pool_counters[p1][x.index].geta(q, BlockRxPipe, pipe, &c.dma_value)
		} else {
			t.rx_pipe_mems.pipe_counter1[p0-3].pool_counters[p1][x.index].geta(q, BlockRxPipe, pipe, &c.dma_value)
		}
	}
	q.Do()

	// Mutually exclude update by fifo dma interrupt.
	pool.mu.Lock()
	c.value.Packets += c.dma_value.Packets
	c.value.Bytes += c.dma_value.Bytes
	v = c.value
	c.value.Zero()
	pool.mu.Unlock()
	return
}

func (t *fe1a) update_pool_counter_values(poolIndex uint, b sbus.Block) {
	fm := &t.pipe_counter_main
	pm := fm.get_pipe(b)

	var pq parallelReq
	q := pq.init(t, 256)

	pool := &pm.pools[poolIndex]
	p0, p1 := poolIndex/4, poolIndex%4
	pool.Foreach(func(pipe, index uint) {
		if index == 0 { // index 0 is always allocated and never valid.
			return
		}
		c := &pool.counters[pipe][index]
		c.dma_value.Zero()
		if b == BlockTxPipe {
			if poolIndex < n_pipe_counter_pool_tx_pipe {
				t.tx_pipe_mems.pipe_counter.pool_counters[poolIndex][index].geta(q, BlockTxPipe, pipe, &c.dma_value)
			} else {
				t.tx_pipe_mems.txf_counter_table[index].geta(q, pipe, &c.dma_value)
			}
		} else {
			if p0 < 3 {
				t.rx_pipe_mems.pipe_counter0[p0].pool_counters[p1][index].geta(q, BlockRxPipe, pipe, &c.dma_value)
			} else {
				t.rx_pipe_mems.pipe_counter1[p0-3].pool_counters[p1][index].geta(q, BlockRxPipe, pipe, &c.dma_value)
			}
		}
		q = pq.do()
	})
	pq.flush()

	// Mutually exclude update by fifo dma interrupt.
	pool.mu.Lock()
	pool.Foreach(func(pipe, index uint) {
		c := &pool.counters[pipe][index]
		if c.dma_value.Packets != 0 {
			c.value.Packets += c.dma_value.Packets
			c.value.Bytes += c.dma_value.Bytes
		}
	})
	pool.mu.Unlock()

	// There might be entries for this pool in eviction fifo.
	t.pipe_counter_eviction_fifo_sync()
}

type pipe_counter_ref_entry struct {
	// Pool index: 2 bits for TX_PIPE.  4 bits for RX_PIPE.
	pool uint8
	// Offset mode: we always set to zero.
	mode uint8
	// Base index of block of counters in counter pool.
	index uint16
	ref   pipemem.Ref
}

func (x *pipe_counter_ref_entry) iGetSet(b []uint32, lo, poolBits, indexBits int, isSet bool) int {
	i := lo
	i = m.MemGetSetUint8(&x.pool, b, i+poolBits, i, isSet)
	i = m.MemGetSetUint8(&x.mode, b, i+1, i, isSet)
	i = m.MemGetSetUint16(&x.index, b, i+indexBits, i, isSet)
	return i
}

type rx_pipe_4p12i_pipe_counter_ref struct{ pipe_counter_ref_entry }

func (x *rx_pipe_4p12i_pipe_counter_ref) MemGetSet(b []uint32, lo int, isSet bool) int {
	return x.iGetSet(b, lo, 4, 12, isSet)
}

type rx_pipe_4p11i_pipe_counter_ref struct{ pipe_counter_ref_entry }

func (x *rx_pipe_4p11i_pipe_counter_ref) MemGetSet(b []uint32, lo int, isSet bool) int {
	return x.iGetSet(b, lo, 4, 11, isSet)
}

type rx_pipe_3p11i_pipe_counter_ref struct{ pipe_counter_ref_entry }

func (x *rx_pipe_3p11i_pipe_counter_ref) MemGetSet(b []uint32, lo int, isSet bool) int {
	return x.iGetSet(b, lo, 3, 11, isSet)
}

func (x *tx_pipe_pipe_counter_ref) MemGetSet(b []uint32, lo int, isSet bool) int {
	i := lo
	i = m.MemGetSetUint16(&x.index, b, i+12, i, isSet)
	i = m.MemGetSetUint8(&x.pool, b, i+3, i, isSet)
	i = m.MemGetSetUint8(&x.mode, b, i+1, i, isSet)
	return i
}

type tx_pipe_pipe_counter_ref struct{ pipe_counter_ref_entry }

type pipe_counter_eviction_fifo_elt struct {
	valid        bool
	counter_wrap bool

	pipe      uint8
	memory_id uint8

	// Evicted counter index.
	index uint16

	packets uint32
	bytes   uint64
}

func (e *pipe_counter_eviction_fifo_elt) pool(pm *pipe_counter_pipe_main, i int) {
	pool := &pm.pools[i]
	pool.mu.Lock()
	c := &pool.counters[e.pipe][e.index]
	c.value.Packets += uint64(e.packets)
	c.value.Bytes += e.bytes
	pool.mu.Unlock()
}

// Indexing from tx_pipe_regs:
// perq_xmit_counters struct {
// 	cpu   [mmu_n_cpu_queues]tx_pipe_pipe_counter_mem
// 	ports [n_idb_mmu_port]struct {
// 		unicast   [mmu_n_tx_queues]tx_pipe_pipe_counter_mem
// 		multicast [mmu_n_tx_queues]tx_pipe_pipe_counter_mem
// 	}
// }
func (e *pipe_counter_eviction_fifo_elt) tx_pipe_perq(t *fe1a) {
	i := e.index
	cm := &t.port_counter_main
	th := t.Vnet.GetIfThread(0)
	if i < mmu_n_cpu_queues {
		k := vnet.HwIfCounterKind(i)
		hi := t.port_by_phys_port[phys_port_cpu].Hi()
		(cm.txPerqCpuCounterKind + 2*k + 0).Add64(th, hi, uint64(e.packets))
		(cm.txPerqCpuCounterKind + 2*k + 1).Add64(th, hi, e.bytes)
	} else {
		i -= mmu_n_cpu_queues
		i0, i1 := i/mmu_n_tx_queues, i%mmu_n_tx_queues
		txq, isMulticast := i1, i0%2 != 0
		pipe_port := pipe_port_number(i0 / 2).mod34_to_pipe(uint(e.pipe))
		phys_port := pipe_port.toPhys()
		hi := t.port_by_phys_port[phys_port].Hi()
		k := vnet.HwIfCounterKind(txq)
		cast := m.Unicast
		if isMulticast {
			cast = m.Multicast
		}
		(cm.txPerqCounterKind[cast] + 2*k + 0).Add64(th, hi, uint64(e.packets))
		(cm.txPerqCounterKind[cast] + 2*k + 1).Add64(th, hi, e.bytes)
	}
}

func (t *fe1a) decode_eviction_fifo(b []uint32) {
	e := pipe_counter_eviction_fifo_elt{}

	{
		isSet := false
		i := 0
		i = m.MemGetSetUint32(&e.packets, b, i+25, i, isSet)
		i = m.MemGetSetUint64(&e.bytes, b, i+33, i, isSet)
		i = m.MemGetSetUint16(&e.index, b, i+12, i, isSet)
		i = m.MemGetSet1(&e.counter_wrap, b, i, isSet)
		i = m.MemGetSet1(&e.valid, b, i, isSet)
		i = m.MemGetSetUint8(&e.memory_id, b, i+5, i, isSet)
		i = m.MemGetSetUint8(&e.pipe, b, i+1, i, isSet)
	}

	if e.counter_wrap {
		panic("pipe counter wrap")
	}

	if !e.valid {
		panic("pipe counter not valid")
	}

	// Find and update counter.
	mi := int(e.memory_id)
	if i, n := pipe_counter_memory_id_pool_rx_pipe, n_pipe_counter_pool_rx_pipe; mi >= i && mi < i+n {
		pm := &t.pipe_counter_main.rx_pipe
		e.pool(pm, mi-i)
	} else if i, n := pipe_counter_memory_id_pool_tx_pipe, n_pipe_counter_pool_tx_pipe; mi >= i && mi < i+n {
		pm := &t.pipe_counter_main.tx_pipe
		e.pool(pm, mi-i)
	} else {
		switch mi {
		case pipe_counter_memory_id_tx_pipe_perq_pool:
			e.tx_pipe_perq(t)
		case pipe_counter_memory_id_tx_pipe_txf_pool:
			pm := &t.pipe_counter_main.tx_pipe
			e.pool(pm, n_pipe_counter_pool_tx_pipe)
		default:
			panic("unknown memory id")
		}
	}
}
