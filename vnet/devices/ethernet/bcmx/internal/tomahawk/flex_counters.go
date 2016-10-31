package tomahawk

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/pipemem"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"

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
	flex_counter_control_enable        = 1 << 1
	flex_counter_control_clear_on_read = 1 << 2
)

// Controls for 4 counter pools.  Only one counter per pool can be incremented per packet.
type flex_counter_4pool_control struct {
	// 2:2 CLR_ON_READ Enables clear-on-read for SBus reads to the attached pool
	// 1:1 COUNTER_POOL_ENABLE 0x1 Control to enable/disable counting per pool.
	// 0:0 MIRROR_COPY_INCR_ENABLE Control to enable/disable counting for mirror copy packets.
	control [4]rx_tx_pipe_reg32

	// Description: 41-to-8 packet attribute masking register for generating packet keys
	// 59:59 USE_COMPRESSED_PKT_KEY If set, compressed packet key is used. To use this, USE_UDF_KEY needs to be set to zero
	// 58:58 USE_UDF_KEY
	// 57:56 USER_SPECIFIED_UDF_VALID User specified UDF valids, these valids along with actual packet udf_valid's are used to increment the counters
	// 55:48 SELECTOR_i_EN Enable for SELECTOR_FOR_BIT_i
	// 47:0 8 6-bit selector for bit i
	packet_attribute_selector_key [4]rx_tx_pipe_reg64

	// 9:8 MODE	R/W Counter eviction mode for the attached pool Encodings are:
	//   0 = Counter eviction is disabled.
	//   1 = Random counter eviction.
	//   2 = Threshold based counter eviction.
	//   3 = Reserved.
	// 7:6 PIPE_ID Identifies the Pipe in which the attached counter pool resides
	// 5:0 MEMORY_ID Unique identifier of a counter memory within a PIPE.
	//   Any pool enabled for counter eviction must be assigned a unique value from [1:CENTRAL_CTR_EVICTION_CONTROL__NUM_CE_PER_PIPE].
	eviction_control [4]rx_tx_pipe_reg32

	// 34:0 SEED Initial seed for LFSR
	eviction_lfsr_seed [4]rx_tx_pipe_reg64

	// 59:26 THRESHOLD_BYTES byte threshold for eviction
	// 25:0 THRESHOLD_PKTS packet threshold for eviction
	eviction_threshold [4]rx_tx_pipe_reg64

	_ [0x80 - 0x14]rx_tx_pipe_reg32

	// undocumented
	offset_table_control [4]rx_tx_pipe_reg32

	_ [0x100 - 0x84]rx_tx_pipe_reg32
}

type flex_counter_1pool_control struct {
	// as above
	control            rx_tx_pipe_reg32
	eviction_control   rx_tx_pipe_reg32
	eviction_threshold rx_tx_pipe_reg64
	eviction_seed      rx_tx_pipe_reg64
}

const (
	n_flex_counter_mode = 4 // 2 bits of mode
)

type flex_counter_4pool_mems struct {
	// Table Min: 0 Table Max: 1023
	// Address: 0x56800000 + j*0x20000 + i*0x1000 i = 0..3 j = 0..2 Block ID: RX_PIPE Access Type: DUPLICATE (9)
	// Description: The actual offset used to index the counter table in the first counter pool is stored in this table
	// 9:9 EVEN_PARITY
	// 8:1 OFFSET This offset will be added to the base index to compute final index
	// 0:0 COUNT_ENABLE Enable/Disable counting for this offset.
	pool_offset_tables [4]struct {
		entries [n_flex_counter_mode][256]m.Mem32
		_       [0x1000 - n_flex_counter_mode*256]m.MemElt
	}

	// Table Min: 0 Table Max: 4095
	// Address: 0x56804000 + j*0x20000 + i*0x4000 i = 0..3 j = 0..2 Block ID: RX_PIPE Access Type: UNIQUE_PIPE0123
	// Description: Counter Table for flexible counter updates. All sources in the Ingress Pipeline including the IFP can update these counters.
	// Index Description: Flex Counter table is derived from offset table and the flex ctr base from each table
	// 67:67 PARITY
	// 66:60 ECC
	// 59:26 BYTE_COUNTER
	// 25:0 PACKET_COUNTER
	pool_counters [4][0x4000]flex_counter_mem

	_ [0x20000 - 0x14000]byte
}

type rx_pipe_flex_counter_mem m.MemElt
type tx_pipe_flex_counter_mem m.MemElt
type flex_counter_mem m.MemElt

type flex_counter_entry struct{ vnet.CombinedCounter }

func (e *flex_counter_entry) MemBits() int { return 68 }
func (e *flex_counter_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint64(&e.Packets, b, i+25, i, isSet)
	i = m.MemGetSetUint64(&e.Bytes, b, i+33, i, isSet)
}

func (e *flex_counter_mem) geta(q *DmaRequest, b sbus.Block, pipe uint, v *flex_counter_entry) {
	(*m.MemElt)(e).MemDmaGet(&q.DmaRequest, v, b, sbus.Unique(pipe))
}
func (e *flex_counter_mem) seta(q *DmaRequest, b sbus.Block, pipe uint, v *flex_counter_entry) {
	(*m.MemElt)(e).MemDmaSet(&q.DmaRequest, v, b, sbus.Unique(pipe))
}

func (e *rx_pipe_flex_counter_mem) geta(q *DmaRequest, rx_pipe uint, v *flex_counter_entry) {
	(*m.MemElt)(e).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, sbus.Unique(rx_pipe))
}
func (e *rx_pipe_flex_counter_mem) seta(q *DmaRequest, rx_pipe uint, v *flex_counter_entry) {
	(*m.MemElt)(e).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, sbus.Unique(rx_pipe))
}
func (e *tx_pipe_flex_counter_mem) geta(q *DmaRequest, tx_pipe uint, v *flex_counter_entry) {
	(*m.MemElt)(e).MemDmaGet(&q.DmaRequest, v, BlockTxPipe, sbus.Unique(tx_pipe))
}
func (e *tx_pipe_flex_counter_mem) seta(q *DmaRequest, tx_pipe uint, v *flex_counter_entry) {
	(*m.MemElt)(e).MemDmaSet(&q.DmaRequest, v, BlockTxPipe, sbus.Unique(tx_pipe))
}

const (
	n_flex_counter_pool_rx_pipe              = 20
	n_flex_counter_pool_tx_pipe              = 4
	flex_counter_memory_id_pool_rx_pipe      = 1
	flex_counter_memory_id_pool_tx_pipe      = flex_counter_memory_id_pool_rx_pipe + n_flex_counter_pool_rx_pipe
	flex_counter_memory_id_tx_pipe_perq_pool = flex_counter_memory_id_pool_tx_pipe + n_flex_counter_pool_tx_pipe
	flex_counter_memory_id_tx_pipe_efp_pool  = flex_counter_memory_id_tx_pipe_perq_pool + 1
	n_flex_counter_memory_id                 = 26
)

func (t *tomahawk) flex_counter_init() {
	q := t.getDmaReq()
	fm := &t.flex_counter_main

	fm.rx_pipe.pools = make([]flex_counter_pool, n_flex_counter_pool_rx_pipe)
	fm.tx_pipe.pools = make([]flex_counter_pool, n_flex_counter_pool_tx_pipe+1) // 1 extra for efp

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

	// Enable threshold based eviction.
	{
		const (
			counter_mode uint32 = 2 << 8 // threshold based eviction
		)

		maxPackets := uint64(1) << 26
		maxBytes := uint64(1) << 34
		nPackets := maxPackets * 3 / 4
		nBytes := maxBytes * 3 / 4
		eviction_threshold := nPackets<<0 | nBytes<<26

		for pipe := uint(0); pipe < n_pipe; pipe++ {
			v := counter_mode | (uint32(pipe) << 6)

			// RxPipe counters.
			mem_id := flex_counter_memory_id_pool_rx_pipe
			for i := range t.rx_pipe_regs.flex_counter {
				for j := range t.rx_pipe_regs.flex_counter[i].eviction_control {
					t.rx_pipe_regs.flex_counter[i].eviction_control[j].seta(q, BlockRxPipe, sbus.Unique(pipe),
						v|uint32(mem_id))
					t.rx_pipe_regs.flex_counter[i].eviction_threshold[j].seta(q, BlockRxPipe, sbus.Unique(pipe),
						eviction_threshold)
					mem_id++
				}
			}

			// Tx_pipe counters.
			t.tx_pipe_regs.perq_counter.eviction_control.seta(q, BlockTxPipe, sbus.Unique(pipe),
				v|flex_counter_memory_id_tx_pipe_perq_pool)
			t.tx_pipe_regs.perq_counter.eviction_threshold.seta(q, BlockTxPipe, sbus.Unique(pipe),
				eviction_threshold)
			t.tx_pipe_regs.efp_counter.eviction_control.seta(q, BlockTxPipe, sbus.Unique(pipe),
				v|flex_counter_memory_id_tx_pipe_efp_pool)
			t.tx_pipe_regs.efp_counter.eviction_threshold.seta(q, BlockTxPipe, sbus.Unique(pipe),
				eviction_threshold)
			for i := range t.tx_pipe_regs.flex_counter.eviction_control {
				t.tx_pipe_regs.flex_counter.eviction_control[i].seta(q, BlockTxPipe, sbus.Unique(pipe),
					v|(flex_counter_memory_id_pool_tx_pipe+uint32(i)))
				t.tx_pipe_regs.flex_counter.eviction_threshold[i].seta(q, BlockTxPipe, sbus.Unique(pipe),
					eviction_threshold)
			}
		}
		q.Do()
	}

	t.flex_counter_fifo_dma_start()

	// Enable central eviction fifo.
	{
		v := uint32(1 << 0) // enable bit
		v |= uint32(n_flex_counter_memory_id << 1)
		t.rx_pipe_regs.flex_counter_eviction_control.set(q, v)
		q.Do()
	}

	// Enable all counters and set to clear on read.
	{
		const v uint32 = flex_counter_control_enable | flex_counter_control_clear_on_read
		t.tx_pipe_regs.perq_counter.control.set(q, BlockTxPipe, v)
		t.tx_pipe_regs.efp_counter.control.set(q, BlockTxPipe, v)
		for i := range t.tx_pipe_regs.flex_counter.control {
			t.tx_pipe_regs.flex_counter.control[i].set(q, BlockTxPipe, v)
		}
		for i := range t.rx_pipe_regs.flex_counter {
			for j := range t.rx_pipe_regs.flex_counter[i].control {
				t.rx_pipe_regs.flex_counter[i].control[j].set(q, BlockRxPipe, v)
			}
		}
		q.Do()
	}
}

func (t *tomahawk) flex_counter_init_offset_table(b sbus.Block, pool uint) {
	q := t.getDmaReq()
	p0, p1 := pool/4, pool%4
	for mode := 0; mode < 4; mode++ {
		for i := 0; i < 256; i++ {
			const enable = 1
			if b == BlockRxPipe {
				if p0 < 3 {
					t.rx_pipe_mems.flex_counter0[p0].pool_offset_tables[p1].entries[mode][i].Set(&q.DmaRequest, b, sbus.Duplicate, enable)
				} else {
					t.rx_pipe_mems.flex_counter1[p0-3].pool_offset_tables[p1].entries[mode][i].Set(&q.DmaRequest, b, sbus.Duplicate, enable)
				}
			} else {
				t.tx_pipe_mems.flex_counter.pool_offset_tables[p1].entries[mode][i].Set(&q.DmaRequest, b, sbus.Duplicate, enable)
			}
		}
		q.Do()
	}
}

// Start fifo dma; must choose non-zero channel.  Zero is reserved for l2 mod fifo.
const flex_counter_fifo_dma_channel = 1

func (t *tomahawk) flex_counter_fifo_dma_start() {
	fm := &t.flex_counter_main

	fm.resultFifo = make(chan sbus.FifoDmaData, 64)
	go func(t *tomahawk) {
		for d := range fm.resultFifo {
			t.handle_fifo_data(d)
		}
	}(t)

	t.Cmic.FifoDmaInit(fm.resultFifo,
		flex_counter_fifo_dma_channel, // channel
		t.rx_pipe_mems.flex_counter_eviction_fifo[0].Address(),
		sbus.Command{
			Opcode:     sbus.FifoPop,
			Block:      BlockRxPipe,
			AccessType: sbus.Unique(0),
			Size:       4,
		},
		// Number of bits in flex counter eviction fifo.
		84,
		// Log2 # of entries in host buffer.
		10)
}

func (t *tomahawk) handle_fifo_data(d sbus.FifoDmaData) {
	if len(d.Data) == 0 {
		return
	}
	fm := &t.flex_counter_main
	// Mutually exclude calls from interrupt and explicit calls via flex_counter_eviction_fifo_sync.
	fm.mu.Lock()
	defer fm.mu.Unlock()
	for i := 0; i < len(d.Data); i += 3 {
		t.decode_eviction_fifo(d.Data[i : i+3])
	}
	d.Free()
}

// Poll eviction fifo for entries.
func (t *tomahawk) flex_counter_eviction_fifo_sync() {
	for {
		d := t.Cmic.FifoDmaSync(flex_counter_fifo_dma_channel)
		if len(d.Data) == 0 {
			break
		}
		t.handle_fifo_data(d)
	}
}

type flex_counter struct {
	ref       pipemem.Ref
	dma_value flex_counter_entry
	value     flex_counter_entry
}

//go:generate gentemplate -d Package=tomahawk -id flex_counter -d VecType=flex_counter_vec -d Type=flex_counter github.com/platinasystems/go/elib/vec.tmpl

type flex_counter_pool struct {
	mu sync.Mutex
	pipemem.Pool
	counters [n_pipe]flex_counter_vec
}

type flex_counter_pipe_main struct {
	pools []flex_counter_pool
}

type flex_counter_main struct {
	rx_pipe, tx_pipe flex_counter_pipe_main
	mu               sync.Mutex
	resultFifo       chan sbus.FifoDmaData
}

func (m *flex_counter_main) get_pipe(b sbus.Block) (pm *flex_counter_pipe_main) {
	pm = &m.rx_pipe
	if b == BlockTxPipe {
		pm = &m.tx_pipe
	}
	return
}

// Pool usage.
const (
	// RxPipe: pools 0-7 have 4k counters; pools 8-19 have 512 counters
	flex_counter_pool_rx_l3_interface = 0
	flex_counter_pool_rx_port_table   = 8
	// Tx_pipe: pools 0-1 have 4k counters; pools 2-3 have 1k counters.
	flex_counter_pool_tx_l3_interface = 0
	flex_counter_pool_tx_adjacency    = 1
	flex_counter_pool_tx_port_table   = 2
)

// Zero index is never valid (enforced by hardware).
func (x *flex_counter_ref_entry) is_valid() bool { return x.index != 0 }

func (x *flex_counter_ref_entry) alloc(t *tomahawk, poolIndex, pipeMask uint, b sbus.Block) (ref pipemem.Ref, ok bool) {
	fm := &t.flex_counter_main
	pm := fm.get_pipe(b)

	pool := &pm.pools[poolIndex]
	if pool.Len() == 0 {
		t.flex_counter_init_offset_table(b, poolIndex)

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

func (x *flex_counter_ref_entry) free(t *tomahawk, b sbus.Block) {
	fm := &t.flex_counter_main
	pm := fm.get_pipe(b)
	pool := &pm.pools[x.pool]
	pool.Put(x.ref)
}

func (x *flex_counter_ref_entry) get_value(t *tomahawk, b sbus.Block) (v vnet.CombinedCounter) {
	fm := &t.flex_counter_main
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

func (x *flex_counter_ref_entry) update_value(t *tomahawk, pipe uint, b sbus.Block) (v flex_counter_entry) {
	fm := &t.flex_counter_main
	pm := fm.get_pipe(b)
	pool := &pm.pools[x.pool]
	c := &pool.counters[pipe][x.index]
	c.dma_value.Zero()
	q := t.getDmaReq()
	if b == BlockTxPipe {
		if x.pool < n_flex_counter_pool_tx_pipe {
			t.tx_pipe_mems.flex_counter.pool_counters[x.pool][x.index].geta(q, BlockTxPipe, pipe, &c.dma_value)
		} else {
			t.tx_pipe_mems.efp_counter_table[x.index].geta(q, pipe, &c.dma_value)
		}
	} else {
		p0, p1 := x.pool/4, x.pool%4
		if p0 < 3 {
			t.rx_pipe_mems.flex_counter0[p0].pool_counters[p1][x.index].geta(q, BlockRxPipe, pipe, &c.dma_value)
		} else {
			t.rx_pipe_mems.flex_counter1[p0-3].pool_counters[p1][x.index].geta(q, BlockRxPipe, pipe, &c.dma_value)
		}
	}
	q.Do()

	// There might be entries for this pool in eviction fifo.
	t.flex_counter_eviction_fifo_sync()

	// Mutually exclude update by fifo dma interrupt.
	pool.mu.Lock()
	c.value.Packets += c.dma_value.Packets
	c.value.Bytes += c.dma_value.Bytes
	v = c.value
	c.value.Zero()
	pool.mu.Unlock()
	return
}

func (t *tomahawk) update_pool_counter_values(poolIndex uint, b sbus.Block) {
	fm := &t.flex_counter_main
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
			if poolIndex < n_flex_counter_pool_tx_pipe {
				t.tx_pipe_mems.flex_counter.pool_counters[poolIndex][index].geta(q, BlockTxPipe, pipe, &c.dma_value)
			} else {
				t.tx_pipe_mems.efp_counter_table[index].geta(q, pipe, &c.dma_value)
			}
		} else {
			if p0 < 3 {
				t.rx_pipe_mems.flex_counter0[p0].pool_counters[p1][index].geta(q, BlockRxPipe, pipe, &c.dma_value)
			} else {
				t.rx_pipe_mems.flex_counter1[p0-3].pool_counters[p1][index].geta(q, BlockRxPipe, pipe, &c.dma_value)
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
	t.flex_counter_eviction_fifo_sync()
}

type flex_counter_ref_entry struct {
	// Pool index: 2 bits for TX_PIPE.  4 bits for RX_PIPE.
	pool uint8
	// Offset mode: we always set to zero.
	mode uint8
	// Base index of block of counters in counter pool.
	index uint16
	ref   pipemem.Ref
}

func (x *flex_counter_ref_entry) iGetSet(b []uint32, lo, poolBits, indexBits int, isSet bool) int {
	i := lo
	i = m.MemGetSetUint8(&x.pool, b, i+poolBits, i, isSet)
	i = m.MemGetSetUint8(&x.mode, b, i+1, i, isSet)
	i = m.MemGetSetUint16(&x.index, b, i+indexBits, i, isSet)
	return i
}

type rx_pipe_4p12i_flex_counter_ref struct{ flex_counter_ref_entry }

func (x *rx_pipe_4p12i_flex_counter_ref) MemGetSet(b []uint32, lo int, isSet bool) int {
	return x.iGetSet(b, lo, 4, 12, isSet)
}

type rx_pipe_4p11i_flex_counter_ref struct{ flex_counter_ref_entry }

func (x *rx_pipe_4p11i_flex_counter_ref) MemGetSet(b []uint32, lo int, isSet bool) int {
	return x.iGetSet(b, lo, 4, 11, isSet)
}

type rx_pipe_3p11i_flex_counter_ref struct{ flex_counter_ref_entry }

func (x *rx_pipe_3p11i_flex_counter_ref) MemGetSet(b []uint32, lo int, isSet bool) int {
	return x.iGetSet(b, lo, 3, 11, isSet)
}

func (x *tx_pipe_flex_counter_ref) MemGetSet(b []uint32, lo int, isSet bool) int {
	i := lo
	i = m.MemGetSetUint16(&x.index, b, i+12, i, isSet)
	i = m.MemGetSetUint8(&x.pool, b, i+3, i, isSet)
	i = m.MemGetSetUint8(&x.mode, b, i+1, i, isSet)
	return i
}

type tx_pipe_flex_counter_ref struct{ flex_counter_ref_entry }

type flex_counter_eviction_fifo_elt struct {
	valid        bool
	counter_wrap bool

	pipe      uint8
	memory_id uint8

	// Evicted counter index.
	index uint16

	packets uint32
	bytes   uint64
}

func (e *flex_counter_eviction_fifo_elt) pool(pm *flex_counter_pipe_main, i int) {
	pool := &pm.pools[i]
	pool.mu.Lock()
	c := &pool.counters[e.pipe][e.index]
	c.value.Packets += uint64(e.packets)
	c.value.Bytes += e.bytes
	pool.mu.Unlock()
}

// Indexing from tx_pipe_regs:
// perq_xmit_counters struct {
// 	cpu   [mmu_n_cpu_queues]tx_pipe_flex_counter_mem
// 	ports [n_idb_mmu_port]struct {
// 		unicast   [mmu_n_tx_queues]tx_pipe_flex_counter_mem
// 		multicast [mmu_n_tx_queues]tx_pipe_flex_counter_mem
// 	}
// }
func (e *flex_counter_eviction_fifo_elt) tx_pipe_perq(t *tomahawk) {
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

func (t *tomahawk) decode_eviction_fifo(b []uint32) {
	e := flex_counter_eviction_fifo_elt{}

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
		panic("flex counter wrap")
	}

	if !e.valid {
		panic("flex counter not valid")
	}

	// Find and update counter.
	mi := int(e.memory_id)
	if i, n := flex_counter_memory_id_pool_rx_pipe, n_flex_counter_pool_rx_pipe; mi >= i && mi < i+n {
		pm := &t.flex_counter_main.rx_pipe
		e.pool(pm, mi-i)
	} else if i, n := flex_counter_memory_id_pool_tx_pipe, n_flex_counter_pool_tx_pipe; mi >= i && mi < i+n {
		pm := &t.flex_counter_main.tx_pipe
		e.pool(pm, mi-i)
	} else {
		switch mi {
		case flex_counter_memory_id_tx_pipe_perq_pool:
			e.tx_pipe_perq(t)
		case flex_counter_memory_id_tx_pipe_efp_pool:
			pm := &t.flex_counter_main.tx_pipe
			e.pool(pm, n_flex_counter_pool_tx_pipe)
		default:
			panic("unknown memory id")
		}
	}
}
