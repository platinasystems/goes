// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
)

const (
	// Service pools 0 through 3.
	n_mmu_service_pool = 4
	// Public pool is index 4.
	n_mmu_service_pool_plus_public = n_mmu_service_pool + 1

	n_mmu_priority_group = 8
)

type mmu_cell_count uint32  // units of 1 cell => 15 bits counts
type mmu_8cell_count uint32 // units of 8 cells => 12 bit counts

func (x *mmu_cell_count) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint32((*uint32)(x), b, i+14, i, isSet)
}
func (x *mmu_8cell_count) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint32((*uint32)(x), b, i+11, i, isSet)
}

type mmu_4mcqe_count uint32 // units of 4 mcqe => 11 bits counts
type mmu_8mcqe_count uint32 // units of 8 mcqes => 10 bit counts

func (x *mmu_4mcqe_count) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint32((*uint32)(x), b, i+10, i, isSet)
}
func (x *mmu_8mcqe_count) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint32((*uint32)(x), b, i+9, i, isSet)
}

type mmu_stats_profile_index uint8

func (x *mmu_stats_profile_index) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+2, i, isSet)
}

type mmu_rx_service_pool_config_entry struct {
	// Number of cells reserved for this port.  Allocated before using cells from main pool.
	min_reserved mmu_cell_count

	// Cell limit for main pool (including min reserved cells).
	// Packets are dropped as cell limit is exceeded.
	limit mmu_cell_count

	// Packets are accepted once cell count drops below resume limit.
	resume_limit mmu_cell_count

	mmu_stats_profile_index
}

func (e *mmu_rx_service_pool_config_entry) MemBits() int { return 55 }
func (e *mmu_rx_service_pool_config_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.min_reserved.MemGetSet(b, i, isSet)
	i = e.limit.MemGetSet(b, i, isSet)
	i = e.resume_limit.MemGetSet(b, i, isSet)
	i = e.mmu_stats_profile_index.MemGetSet(b, i, isSet)
}

type mmu_rx_service_pool_config_mem m.MemElt

func (e *mmu_rx_service_pool_config_mem) get(q *DmaRequest, v *mmu_rx_service_pool_config_entry) {
	(*m.MemElt)(e).MemDmaGeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeRxPipe)
}
func (e *mmu_rx_service_pool_config_mem) set(q *DmaRequest, v *mmu_rx_service_pool_config_entry) {
	(*m.MemElt)(e).MemDmaSeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeRxPipe)
}

// Up to 8 priority groups (mapped from packet internal priority) may share a single service pool.
type mmu_priority_group_config_entry struct {
	// Number of cells reserved for this port.  Allocated before using cells from shared pool.
	min_reserved mmu_cell_count

	shared_limit_is_dynamic bool

	// 2^(X-7) fraction of current total available shared cells.
	log2_dynamic_shared_limit uint8

	// Cell limit for shared pool (including min reserved cells).
	// Only used when dynamic mode is false.
	shared_limit mmu_cell_count

	// If set this priority may use global headroom; otherwise only this priority group's headroom.
	global_headroom_enable bool

	// Number of headroom cells this priority group may use.
	headroom_limit mmu_cell_count

	// Specify XON/XOFF
	// Below floor => release xoff state; above floor + offset => enter xoff state.
	reset_floor  mmu_cell_count
	reset_offset mmu_cell_count

	shared_stats_profile_index   mmu_stats_profile_index
	headroom_stats_profile_index mmu_stats_profile_index
}

func (e *mmu_priority_group_config_entry) MemBits() int { return 91 }
func (e *mmu_priority_group_config_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.min_reserved.MemGetSet(b, i, isSet)
	v := e.shared_limit
	if isSet {
		if e.shared_limit_is_dynamic {
			v = mmu_cell_count(e.log2_dynamic_shared_limit)
		}
	}
	i = v.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.shared_limit_is_dynamic, b, i, isSet)
	if !isSet {
		if e.shared_limit_is_dynamic {
			e.shared_limit = 0
			e.log2_dynamic_shared_limit = uint8(v)
		} else {
			e.shared_limit = v
			e.log2_dynamic_shared_limit = 0
		}
	}
	i = e.headroom_limit.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.global_headroom_enable, b, i, isSet)
	i = e.reset_offset.MemGetSet(b, i, isSet)
	i = e.reset_floor.MemGetSet(b, i, isSet)
	i = e.shared_stats_profile_index.MemGetSet(b, i, isSet)
	i = e.headroom_stats_profile_index.MemGetSet(b, i, isSet)
}

type mmu_priority_group_config_mem m.MemElt

func (e *mmu_priority_group_config_mem) get(q *DmaRequest, v *mmu_priority_group_config_entry) {
	(*m.MemElt)(e).MemDmaGeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeRxPipe)
}
func (e *mmu_priority_group_config_mem) set(q *DmaRequest, v *mmu_priority_group_config_entry) {
	(*m.MemElt)(e).MemDmaSeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeRxPipe)
}

type mmu_tx_service_pool_config_entry struct {
	shared_limit mmu_cell_count
	yellow_limit mmu_8cell_count
	red_limit    mmu_8cell_count
}

func (e *mmu_tx_service_pool_config_entry) MemBits() int { return 49 }
func (e *mmu_tx_service_pool_config_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.shared_limit.MemGetSet(b, i, isSet)
	i = e.yellow_limit.MemGetSet(b, i, isSet)
	i = e.red_limit.MemGetSet(b, i, isSet)
}

type mmu_tx_service_pool_config_mem m.MemElt

func (e *mmu_tx_service_pool_config_mem) get(q *DmaRequest, v *mmu_tx_service_pool_config_entry) {
	(*m.MemElt)(e).MemDmaGeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeTxPipe)
}
func (e *mmu_tx_service_pool_config_mem) set(q *DmaRequest, v *mmu_tx_service_pool_config_entry) {
	(*m.MemElt)(e).MemDmaSeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeTxPipe)
}

type mmu_tx_color_config_entry struct {
	green  mmu_8cell_count
	yellow mmu_8cell_count
	red    mmu_8cell_count
}

func (e *mmu_tx_color_config_entry) MemBits() int { return 43 }
func (e *mmu_tx_color_config_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.green.MemGetSet(b, i, isSet)
	i = e.yellow.MemGetSet(b, i, isSet)
	i = e.red.MemGetSet(b, i, isSet)
}

type mmu_tx_color_config_mem m.MemElt

func (e *mmu_tx_color_config_mem) get(q *DmaRequest, v *mmu_tx_color_config_entry) {
	(*m.MemElt)(e).MemDmaGeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeTxPipe)
}
func (e *mmu_tx_color_config_mem) set(q *DmaRequest, v *mmu_tx_color_config_entry) {
	(*m.MemElt)(e).MemDmaSeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeTxPipe)
}

type mmu_tx_queue_config_entry struct {
	// Number of cells reserved for this queue.
	min_reserved mmu_cell_count

	shared_limit_is_dynamic bool

	// 2^(X-7) fraction of current total available shared cells.
	log2_dynamic_shared_limit uint8

	// Cell limit for shared pool (including min reserved cells).
	// Only used when dynamic mode is false.
	shared_limit mmu_cell_count

	color_limit_is_dynamic bool

	// 0 => 100% else in units of l/8 of green threshold
	red_limit    mmu_8cell_count
	yellow_limit mmu_8cell_count

	// BST threshold index for this queue.
	bst_max_threshold mmu_stats_profile_index
}

func (e *mmu_tx_queue_config_entry) MemBits() int { return 70 }
func (e *mmu_tx_queue_config_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	v := e.shared_limit
	if isSet {
		if e.shared_limit_is_dynamic {
			v = mmu_cell_count(e.log2_dynamic_shared_limit)
		}
	}
	i = v.MemGetSet(b, i, isSet)
	i = e.min_reserved.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.shared_limit_is_dynamic, b, i, isSet)
	if !isSet {
		if e.shared_limit_is_dynamic {
			e.shared_limit = 0
			e.log2_dynamic_shared_limit = uint8(v)
		} else {
			e.shared_limit = v
			e.log2_dynamic_shared_limit = 0
		}
	}

	i = m.MemGetSet1(&e.color_limit_is_dynamic, b, i, isSet)
	i = e.yellow_limit.MemGetSet(b, i, isSet)
	i = e.red_limit.MemGetSet(b, i, isSet)
	i = e.bst_max_threshold.MemGetSet(b, i, isSet)
}

type mmu_tx_queue_config_mem m.MemElt

func (e *mmu_tx_queue_config_mem) get(q *DmaRequest, v *mmu_tx_queue_config_entry) {
	(*m.MemElt)(e).MemDmaGeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeTxPipe)
}
func (e *mmu_tx_queue_config_mem) set(q *DmaRequest, v *mmu_tx_queue_config_entry) {
	(*m.MemElt)(e).MemDmaSeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeTxPipe)
}

type mmu_tx_queue_to_queue_group_map_entry struct {
	queue_group_valid bool

	// False => queue uses its own min space; other queues uses min space from group.
	use_queue_group_min_threshold bool

	// When true packets for this queue are unconditionally dropped.
	disable bool

	use_mop1b_ticket bool

	queue_color_discard_threshold_enable bool

	queue_limit_enable       bool
	queue_group_limit_enable bool

	service_pool uint8
}

func (e *mmu_tx_queue_to_queue_group_map_entry) MemBits() int { return 25 }
func (e *mmu_tx_queue_to_queue_group_map_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSet1(&e.queue_group_valid, b, i, isSet)
	i = m.MemGetSet1(&e.use_queue_group_min_threshold, b, i, isSet)
	i = m.MemGetSet1(&e.disable, b, i, isSet)
	i = m.MemGetSet1(&e.use_mop1b_ticket, b, i, isSet)
	i = m.MemGetSetUint8(&e.service_pool, b, i+1, i, isSet)
	i = m.MemGetSet1(&e.queue_color_discard_threshold_enable, b, i, isSet)
	i = 9
	i = m.MemGetSet1(&e.queue_limit_enable, b, i, isSet)
	i = m.MemGetSet1(&e.queue_group_limit_enable, b, i, isSet)
}

type mmu_tx_queue_to_queue_group_map_mem m.MemElt

func (e *mmu_tx_queue_to_queue_group_map_mem) get(q *DmaRequest, v *mmu_tx_queue_to_queue_group_map_entry) {
	(*m.MemElt)(e).MemDmaGeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeTxPipe)
}
func (e *mmu_tx_queue_to_queue_group_map_mem) set(q *DmaRequest, v *mmu_tx_queue_to_queue_group_map_entry) {
	(*m.MemElt)(e).MemDmaSeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeTxPipe)
}

type mmu_multicast_db_service_pool_config_entry struct {
	shared_limit_enable bool
	shared_limit        mmu_cell_count
	shared_resume_limit mmu_8cell_count

	yellow_shared_limit mmu_8cell_count
	yellow_resume_limit mmu_8cell_count
	red_shared_limit    mmu_8cell_count
	red_resume_limit    mmu_8cell_count
}

func (e *mmu_multicast_db_service_pool_config_entry) MemBits() int { return 84 }
func (e *mmu_multicast_db_service_pool_config_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.shared_limit.MemGetSet(b, i, isSet)
	i = e.shared_resume_limit.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.shared_limit_enable, b, i, isSet)
	i = e.yellow_shared_limit.MemGetSet(b, i, isSet)
	i = e.red_shared_limit.MemGetSet(b, i, isSet)
	i = e.yellow_resume_limit.MemGetSet(b, i, isSet)
	i = e.red_resume_limit.MemGetSet(b, i, isSet)
	if i != 76 {
		panic("76")
	}
}

type mmu_multicast_db_service_pool_config_mem m.MemElt

func (e *mmu_multicast_db_service_pool_config_mem) get(q *DmaRequest, v *mmu_multicast_db_service_pool_config_entry) {
	(*m.MemElt)(e).MemDmaGeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeTxPipe)
}
func (e *mmu_multicast_db_service_pool_config_mem) set(q *DmaRequest, v *mmu_multicast_db_service_pool_config_entry) {
	(*m.MemElt)(e).MemDmaSeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeTxPipe)
}

type mmu_multicast_mcqe_service_pool_config_entry struct {
	shared_limit_enable bool
	shared_limit        mmu_4mcqe_count
	shared_resume_limit mmu_8mcqe_count

	yellow_shared_limit mmu_8mcqe_count
	yellow_resume_limit mmu_8mcqe_count
	red_shared_limit    mmu_8mcqe_count
	red_resume_limit    mmu_8mcqe_count
}

func (e *mmu_multicast_mcqe_service_pool_config_entry) MemBits() int { return 84 }
func (e *mmu_multicast_mcqe_service_pool_config_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.shared_limit.MemGetSet(b, i, isSet)
	i = e.shared_resume_limit.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.shared_limit_enable, b, i, isSet)
	i = e.yellow_shared_limit.MemGetSet(b, i, isSet)
	i = e.red_shared_limit.MemGetSet(b, i, isSet)
	i = e.yellow_resume_limit.MemGetSet(b, i, isSet)
	i = e.red_resume_limit.MemGetSet(b, i, isSet)
	if i != 62 {
		panic("62")
	}
}

type mmu_multicast_mcqe_service_pool_config_mem m.MemElt

func (e *mmu_multicast_mcqe_service_pool_config_mem) get(q *DmaRequest, v *mmu_multicast_mcqe_service_pool_config_entry) {
	(*m.MemElt)(e).MemDmaGeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeTxPipe)
}
func (e *mmu_multicast_mcqe_service_pool_config_mem) set(q *DmaRequest, v *mmu_multicast_mcqe_service_pool_config_entry) {
	(*m.MemElt)(e).MemDmaSeta(&q.DmaRequest, v, BlockMmuXpe, sbus.Duplicate, mmuBaseTypeTxPipe)
}

const (
	// Packets are split into fixed sized cells.
	// First cell in packet has 64 byte header/meta data + 144 bytes of packet data.
	// Subsequent cells are full 208 bytes of packet data.
	n_mmu_cell_bytes                       = 208
	n_mmu_first_packet_cell_overhead_bytes = 64

	// Packet memory has 15 banks of 1536 cells
	n_mmu_cell_buffer_banks          = 15
	n_mmu_cell_buffer_slices         = 4
	n_mmu_cell_buffer_cells_per_bank = 1536
)

type mmu_cell_buffer_pool_entry [14]uint32

func (e *mmu_cell_buffer_pool_entry) MemBits() int { return 419 }
func (e *mmu_cell_buffer_pool_entry) MemGetSet(b []uint32, isSet bool) {
	if isSet {
		copy(b, e[:])
	} else {
		copy(e[:], b)
	}
}

type mmu_cell_buffer_pool_mem m.MemElt

func (e *mmu_cell_buffer_pool_mem) get(q *DmaRequest, v *mmu_cell_buffer_pool_entry) {
	(*m.MemElt)(e).MemDmaGeta(&q.DmaRequest, v, BlockMmuSc, sbus.Single, mmuBaseTypeXpe)
}
func (e *mmu_cell_buffer_pool_mem) set(q *DmaRequest, v *mmu_cell_buffer_pool_entry) {
	(*m.MemElt)(e).MemDmaSeta(&q.DmaRequest, v, BlockMmuSc, sbus.Single, mmuBaseTypeXpe)
}
