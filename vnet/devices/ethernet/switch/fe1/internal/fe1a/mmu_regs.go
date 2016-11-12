// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
)

const (
	n_mmu_pipe              = 4
	n_mmu_slice_controllers = 2
	n_mmu_layer             = 2
	mmu_per_pipe_mem_bytes  = 0x8000
	mmu_n_cells_per_bank    = 1536
	mmu_n_banks             = 15
	mmu_n_cells_per_xpe     = mmu_n_banks * mmu_n_cells_per_bank

	mmu_per_packet_metadata_overhead = 64

	mmu_cell_bytes = 208

	mmu_reserved_cfap_cells = 0

	mmu_max_packet_bytes = 9416

	mmu_n_tx_cos_queues = 8
	mmu_n_tx_queues     = mmu_n_tx_cos_queues + 2
	mmu_n_cpu_queues    = 48

	mmu_n_rqe_queue = mmu_n_tx_cos_queues + 2 + 1

	mmu_n_mcqe = 8 << 10
	mmu_n_rqe  = 1 << 10
)

func tx_pipe_xpe_mask(tx_pipe uint) (mask uint) {
	switch tx_pipe {
	case 0, 1:
		mask = 1<<0 | 1<<2
	case 2, 3:
		mask = 1<<1 | 1<<3
	default:
		panic("tx pipe")
	}
	return
}

func rx_pipe_xpe_mask(rx_pipe uint) (mask uint) {
	switch rx_pipe {
	case 0, 3:
		mask = 1<<0 | 1<<1
	case 1, 2:
		mask = 1<<2 | 1<<3
	default:
		panic("rx pipe")
	}
	return
}

var mmu_tx_queue_names = [...]string{"cos0", "cos1", "cos2", "cos3", "cos4", "cos5", "cos6", "cos7", "sc", "qm"}

func bytesToCells(b uint) uint {
	return (b + mmu_per_packet_metadata_overhead + mmu_cell_bytes - 1) / mmu_cell_bytes
}

const (
	mmuBaseTypeRxPort sbus.Address = iota << 23
	mmuBaseTypeTxPort
	mmuBaseTypeRxPipe
	mmuBaseTypeTxPipe
	mmuBaseTypeChip
	mmuBaseTypeXpe
	mmuBaseTypeSlice
	mmuBaseTypeLayer
)

type mmu_global_reg32 m.U32

func (r *mmu_global_reg32) geta(q *DmaRequest, c sbus.AccessType, v *uint32) {
	(*m.U32)(r)[0].Get(&q.DmaRequest, mmuBaseTypeChip, BlockMmuGlobal, c, v)
}
func (r *mmu_global_reg32) seta(q *DmaRequest, c sbus.AccessType, v uint32) {
	(*m.U32)(r)[0].Set(&q.DmaRequest, mmuBaseTypeChip, BlockMmuGlobal, c, v)
}

func (r *mmu_global_reg32) get(q *DmaRequest, v *uint32) { r.geta(q, sbus.Single, v) }
func (r *mmu_global_reg32) set(q *DmaRequest, v uint32)  { r.seta(q, sbus.Single, v) }

func (r *mmu_global_reg32) getDo(q *DmaRequest) (v uint32) {
	r.get(q, &v)
	q.Do()
	return v
}

type mmu_global_reg64 m.U64

func (r *mmu_global_reg64) geta(q *DmaRequest, c sbus.AccessType, v *uint64) {
	(*m.U64)(r).Get(&q.DmaRequest, mmuBaseTypeChip, BlockMmuGlobal, c, v)
}
func (r *mmu_global_reg64) seta(q *DmaRequest, c sbus.AccessType, v uint64) {
	(*m.U64)(r).Set(&q.DmaRequest, mmuBaseTypeChip, BlockMmuGlobal, c, v)
}

func (r *mmu_global_reg64) get(q *DmaRequest, v *uint64) { r.geta(q, sbus.Single, v) }
func (r *mmu_global_reg64) set(q *DmaRequest, v uint64)  { r.seta(q, sbus.Single, v) }

type mmu_global_preg32 m.Pu32
type mmu_global_portreg32 [1 << m.Log2NPorts]mmu_global_preg32

func (r *mmu_global_preg32) geta(q *DmaRequest, c sbus.AccessType, v *uint32) {
	(*m.Pu32)(r).Get(&q.DmaRequest, mmuBaseTypeRxPort, BlockMmuGlobal, c, v)
}
func (r *mmu_global_preg32) seta(q *DmaRequest, c sbus.AccessType, v uint32) {
	(*m.Pu32)(r).Set(&q.DmaRequest, mmuBaseTypeRxPort, BlockMmuGlobal, c, v)
}
func (r *mmu_global_preg32) set(q *DmaRequest, v uint32) { r.seta(q, sbus.Single, v) }

type mmu_global_controller struct {
	_ [0x08000000 - 0x0]byte

	misc_config mmu_global_reg32

	_ [1]m.U32

	parity_interrupt struct {
		enable                  mmu_global_reg32
		status                  mmu_global_reg32
		status_write_1_to_clear mmu_global_reg32
	}

	interrupt_enable mmu_global_reg32

	interrupt_status mmu_global_reg32

	_ [0xb - 0x7]m.U32

	bst_tracking_enable mmu_global_reg32

	bst_hw_snapshot_reset mmu_global_reg32

	bst_hw_snapshot_enable mmu_global_reg32

	_ [0x1000 - 0xe]m.U32

	device_port_by_mmu_port mmu_global_portreg32

	_ [0x1100 - 0x1001]m.U32

	physical_port_by_mmu_port mmu_global_portreg32

	_ [0x1200 - 0x1101]m.U32

	global_physical_port_by_mmu_port mmu_global_portreg32
}

type mmu_global_mems struct{}

type mmu_pipe_reg32 [1 << m.Log2NPorts]m.Gu32

func (r *mmu_pipe_reg32) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint32) {
	(*m.Gu32)(&r[0]).Get(&q.DmaRequest, a, BlockMmuXpe, c, v)
}
func (r *mmu_pipe_reg32) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint32) {
	(*m.Gu32)(&r[0]).Set(&q.DmaRequest, a, BlockMmuXpe, c, v)
}

type mmu_pipe_gpreg32 m.Pu32
type mmu_pipe_xreg32 [1 << m.Log2NPorts]mmu_pipe_gpreg32

func (r *mmu_pipe_gpreg32) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint32) {
	(*m.Pu32)(r).Get(&q.DmaRequest, a, BlockMmuXpe, c, v)
}
func (r *mmu_pipe_gpreg32) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint32) {
	(*m.Pu32)(r).Set(&q.DmaRequest, a, BlockMmuXpe, c, v)
}

type mmu_pipe_gpreg64 m.Pu64
type mmu_pipe_xreg64 [1 << m.Log2NPorts]mmu_pipe_gpreg64

func (r *mmu_pipe_gpreg64) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint64) {
	(*m.Pu64)(r).Get(&q.DmaRequest, a, BlockMmuXpe, c, v)
}
func (r *mmu_pipe_gpreg64) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint64) {
	(*m.Pu64)(r).Set(&q.DmaRequest, a, BlockMmuXpe, c, v)
}

type mmu_pipe_preg32 m.Pu32
type mmu_pipe_portreg32 [1 << m.Log2NPorts]mmu_pipe_preg32

func (r *mmu_pipe_preg32) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint32) {
	(*m.Pu32)(r).Get(&q.DmaRequest, a, BlockMmuXpe, c, v)
}
func (r *mmu_pipe_preg32) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint32) {
	(*m.Pu32)(r).Set(&q.DmaRequest, a, BlockMmuXpe, c, v)
}

type mmu_pipe_regs struct {
	_ [0x08000000 - 0x0]byte

	rx_admission_control struct {
		port_rx_and_pause_enable mmu_pipe_portreg32

		_ [0x7 - 0x1]m.U32

		global_headroom_limit mmu_pipe_xreg32

		_ [0x22 - 0x8]m.U32

		max_packet_cells mmu_pipe_reg32

		_ [0x70 - 0x23]m.U32

		port_priority_group [2]mmu_pipe_portreg32

		_ [0x73 - 0x72]m.U32

		port_service_pool_for_priority_group mmu_pipe_portreg32

		port_headroom_pool_for_priority_group mmu_pipe_portreg32

		_ [0x7b - 0x75]m.U32

		global_headroom_cells_in_use          mmu_pipe_xreg32
		global_headroom_reserved_cells_in_use mmu_pipe_xreg32

		_ [0x100 - 0x7d]m.U32

		service_pool_n_shared_cells_in_use [n_mmu_service_pool_plus_public]mmu_pipe_xreg32

		_ [0x10a - 0x105]m.U32

		service_pool_shared_cell_limit [n_mmu_service_pool_plus_public]mmu_pipe_xreg32

		_ [0x114 - 0x10f]m.U32

		service_pool_cell_reset_limit_offset [n_mmu_service_pool]mmu_pipe_xreg32

		_ [0x11c - 0x118]m.U32

		cell_spap_yellow_offset [n_mmu_service_pool]mmu_pipe_xreg32
		cell_spap_red_offset    [n_mmu_service_pool]mmu_pipe_xreg32

		_ [0x131 - 0x124]m.U32

		service_pool_config mmu_pipe_xreg32

		_ [0x134 - 0x132]m.U32

		port_limit_states mmu_pipe_portreg32

		_ [0x138 - 0x135]m.U32

		priority_group_xoff_status mmu_pipe_portreg32

		pool_drop_state mmu_pipe_xreg32

		_ [0x140 - 0x13a]m.U32

		headroom_pool_peak_count_update_enable mmu_pipe_xreg32

		headroom_pool_status mmu_pipe_xreg32

		_ [0x150 - 0x142]m.U32

		headroom_pool_n_cells_in_use [4]mmu_pipe_xreg32

		headroom_pool_cell_limit [4]mmu_pipe_xreg32

		headroom_pool_peak_cell_count [4]mmu_pipe_xreg32

		_ [0x170 - 0x15c]m.U32

		buffer_stats_tracking struct {
			priority_group_shared_cell_threshold   [8]mmu_pipe_reg32
			priority_group_headroom_cell_threshold [8]mmu_pipe_reg32
			service_pool_shared_cell_threshold     [8]mmu_pipe_reg32

			_ [0x190 - 0x188]m.U32

			service_pool_global_shared_cells [n_mmu_service_pool_plus_public]mmu_pipe_reg32

			_ [8 - n_mmu_service_pool_plus_public]mmu_pipe_reg32

			service_pool_global_shared_cell_threshold [8]mmu_pipe_xreg32

			trigger_status_type mmu_pipe_xreg32

			trigger_status mmu_pipe_xreg32
		}
	}

	_ [0x10002b00 - 0x0801a200]byte

	cut_through_purge_count mmu_pipe_xreg32

	_ [0x14001000 - 0x10002c00]byte

	ccp_status mmu_pipe_xreg32

	_ [0x20000000 - 0x14001100]byte

	time_domain [4]mmu_pipe_reg32

	_ [0x6 - 0x4]m.U32

	wred_refresh_control mmu_pipe_reg32

	wred_congestion_notification_resolution_table [4]mmu_pipe_reg32

	wred_pool_congestion_limit [4]mmu_pipe_reg32

	_ [0x24000000 - 0x20000f00]byte

	pqe [2]struct {
		fifo_empty_port_bitmap [2]mmu_pipe_xreg32

		fifo_overflow_port_bitmap [2]mmu_pipe_xreg32

		_ [0x5 - 0x4]m.U32

		fifo_pointer_equal_bitmap [2]mmu_pipe_xreg32

		_ [0x8 - 0x7]m.U32

		qcn_fifo_empty_port_bitmap [2]mmu_pipe_xreg32

		_ [0x10 - 0xa]m.U32
	}

	_ [0x28000000 - 0x24002000]byte

	clear_counters mmu_pipe_reg32

	_ [0x10 - 0x1]m.U32

	tx_counter_config [8]mmu_pipe_xreg32

	_ [0x20 - 0x18]m.U32

	tx_counter_packets [8]mmu_pipe_xreg64

	_ [0x110 - 0x28]m.U32

	multicast_replication_fifo_drop_packets [mmu_n_rqe_queue]mmu_pipe_xreg64

	_ [0x120 - 0x11b]m.U32

	multicast_replication_fifo_drop_bytes [mmu_n_rqe_queue]mmu_pipe_xreg64
	_                                     [0x12f - 0x12b]m.U32

	data_buffer_full_packets_dropped mmu_pipe_xreg64

	_ [0x200 - 0x130]m.U32

	multicast_replication_fifo_red_drop_packets [mmu_n_rqe_queue]mmu_pipe_xreg64

	_ [0x220 - 0x20b]m.U32

	multicast_replication_fifo_yellow_drop_packets [mmu_n_rqe_queue]mmu_pipe_xreg64

	_ [0x2c003600 - 0x28022b00]byte

	rqe_priority_scheduling_type [mmu_n_rqe_queue]mmu_pipe_reg32

	_ [0x60 - 0x41]m.U32

	rqe_priority_werr_weight [mmu_n_rqe_queue]mmu_pipe_reg32

	_ [0x88 - 0x6b]m.U32

	port_cos_mode mmu_pipe_xreg64

	_ [0x110 - 0x89]m.U32

	port_initial_copy_count_width mmu_pipe_portreg32

	_ [0x38001000 - 0x2c011100]byte

	tx_admission_control struct {
		bypass mmu_pipe_reg32

		_ [0x12 - 0x11]m.U32

		config mmu_pipe_reg32

		_ [0x74 - 0x13]m.U32

		tx_port_enable mmu_pipe_xreg64

		_ [0x100 - 0x75]m.U32

		queue_drop_state_bitmap [6]mmu_pipe_xreg64

		_ [0x200 - 0x106]m.U32

		queue_yellow_drop_state_bitmap [6]mmu_pipe_xreg64
		_                              [0x300 - 0x206]m.U32

		queue_red_drop_state_bitmap [6]mmu_pipe_xreg64

		_ [0x500 - 0x306]m.U32

		queue_group_drop_state mmu_pipe_xreg64

		_ [0x520 - 0x501]m.U32

		queue_group_yellow_drop_state mmu_pipe_xreg64

		_ [0x540 - 0x521]m.U32

		queue_group_red_drop_state mmu_pipe_xreg64

		_ [0x610 - 0x541]m.U32

		per_service_pool_port_drop_state [4]mmu_pipe_xreg64

		_ [0x620 - 0x614]m.U32

		per_service_pool_port_yellow_drop_state [4]mmu_pipe_xreg64

		_ [0x630 - 0x624]m.U32

		per_service_pool_port_red_drop_state [4]mmu_pipe_xreg64

		_ [0x820 - 0x634]m.U32

		queue_bst_threshold [8]mmu_pipe_xreg32

		queue_group_bst_threshold [8]mmu_pipe_xreg32

		port_bst_threshold [8]mmu_pipe_xreg32

		_ [0x850 - 0x838]m.U32

		bst_status mmu_pipe_xreg32

		_ [0x900 - 0x851]m.U32

		port_e2ecc_cos_spid mmu_pipe_portreg32

		_ [0xa0b - 0x901]m.U32

		congestion_state_reset mmu_pipe_xreg32
	}

	_ [0x3c000400 - 0x380a0c00]byte

	multicast_admission_control struct {
		db struct {
			service_pool_shared_limit [4]mmu_pipe_reg32

			service_pool_yellow_shared_limit [4]mmu_pipe_reg32
			service_pool_red_shared_limit    [4]mmu_pipe_reg32

			_ [0x12 - 0x10]m.U32

			config mmu_pipe_reg32

			_ [0x30 - 0x13]m.U32

			service_pool_bst_threshold_profile_select [4]mmu_pipe_portreg32

			_ [0x40 - 0x34]m.U32

			service_pool_shared_count [4]mmu_pipe_xreg32

			_ [0x4c - 0x44]m.U32

			service_pool_multicast_shared_count [4]mmu_pipe_xreg32

			port_service_pool_shared_count [4]mmu_pipe_portreg32

			_ [0x6c - 0x54]m.U32

			service_pool_drop_states mmu_pipe_xreg32

			_ [0x74 - 0x6d]m.U32

			port_tx_enable mmu_pipe_xreg64

			_ [0x80 - 0x75]m.U32

			port_service_pool_yellow_drop_state [4]mmu_pipe_xreg64

			service_pool_resume_limit [n_packet_color][4]mmu_pipe_reg32

			_ [0xa0 - 0x90]m.U32

			port_service_pool_red_drop_state [4]mmu_pipe_xreg64

			_ [0x260 - 0xa4]m.U32

			port_service_pool_drop_state [4]mmu_pipe_xreg64

			_ [0x500 - 0x264]m.U32

			cpu_queue_bst_threshold_profile [8]mmu_pipe_xreg32

			_ [0x540 - 0x508]m.U32

			queue_resume_offset_profile_yellow [8]mmu_pipe_reg32

			queue_resume_offset_profile_red [8]mmu_pipe_reg32

			queue_e2e_ds mmu_pipe_portreg32

			_ [0x638 - 0x551]m.U32

			service_pool_mcuc_bst_threshold [4]mmu_pipe_reg32

			service_pool_mcuc_bst_stat [4]mmu_pipe_xreg32

			queue_multicast_bst_threshold_profile [8]mmu_pipe_reg32

			_ [0x650 - 0x648]m.U32

			device_bst_status mmu_pipe_xreg64

			_ [0x6a0 - 0x651]m.U32

			service_pool_multicast_bst_threshold [4]mmu_pipe_reg32

			service_pool_multicast_bst_status [4]mmu_pipe_xreg32

			_ [0x6b0 - 0x6a8]m.U32

			port_service_pool_bst_threshold [8]mmu_pipe_reg32

			_ [0x700 - 0x6b8]m.U32

			queue_e2e_spid mmu_pipe_portreg32

			_ [0x950 - 0x701]m.U32

			enable_eccp_mem mmu_pipe_reg32

			enable_correctable_error_reporting mmu_pipe_reg32

			_ [0xa00 - 0x952]m.U32

			queue_e2e_ds_en mmu_pipe_portreg32

			_ [0x40000400 - 0x3c0a0100]byte
		}

		mcqe struct {
			pool_shared_limit [4]mmu_pipe_reg32

			pool_yellow_shared_limit [4]mmu_pipe_reg32
			pool_red_shared_limit    [4]mmu_pipe_reg32

			_ [0x12 - 0x10]m.U32

			config mmu_pipe_reg32

			_ [0x30 - 0x13]m.U32

			port_service_pool_bst_threshold_profile_select [4]mmu_pipe_xreg32

			_ [0x40 - 0x34]m.U32

			pool_shared_count [4]mmu_pipe_xreg32

			_ [0x50 - 0x44]m.U32

			port_service_pool_shared_count [4]mmu_pipe_portreg32

			_ [0x6c - 0x54]m.U32

			pool_drop_states mmu_pipe_xreg32

			_ [0x74 - 0x6d]m.U32

			port_tx_enable mmu_pipe_xreg64

			_ [0x80 - 0x75]m.U32

			port_service_pool_yellow_drop_state [4]mmu_pipe_xreg64

			pool_resume_limit [n_packet_color][4]mmu_pipe_reg32

			_ [0xa0 - 0x90]m.U32

			port_service_pool_red_drop_state [4]mmu_pipe_xreg64

			_ [0x260 - 0xa4]m.U32

			port_service_pool_drop_state [4]mmu_pipe_xreg64

			_ [0x500 - 0x264]m.U32

			cpu_queue_bst_threshold_profile [8]mmu_pipe_xreg32

			_ [0x540 - 0x508]m.U32

			queue_resume_offset_profile_yellow [8]mmu_pipe_reg32
			queue_resume_offset_profile_red    [8]mmu_pipe_reg32

			_ [0x640 - 0x550]m.U32

			queue_bst_threshold_profile [8]mmu_pipe_reg32

			_ [0x650 - 0x648]m.U32

			bst_status mmu_pipe_xreg32

			_ [0x6a0 - 0x651]m.U32

			pool_multicast_bst_threshold [4]mmu_pipe_reg32

			pool_multicast_bst_status [4]mmu_pipe_xreg32

			_ [0x6b0 - 0x6a8]m.U32

			port_service_pool_bst_threshold [8]mmu_pipe_reg32

			_ [0x850 - 0x6b8]m.U32

			enable_ecc_parity mmu_pipe_reg32

			enable_ecc_error_reporting mmu_pipe_reg32
		}
	}

	_ [0x44000000 - 0x40085200]byte

	db mmu_pipe_rqe_admission_control_regs

	_ [0x48010000 - 0x44010000]byte

	qe mmu_pipe_rqe_admission_control_regs

	_ [0x4c000100 - 0x48020000]byte

	interrupt_enable mmu_pipe_reg32
	interrupt_status mmu_pipe_xreg32
	intterupt_clear  mmu_pipe_xreg32
}

type mmu_pipe_rqe_admission_control_regs struct {
	_ [1]m.U32

	config mmu_pipe_reg32

	service_pool_config [4]mmu_pipe_reg32

	service_pool_per_color_shared_limit [4]mmu_pipe_reg32

	resume_color_limit [4]mmu_pipe_reg32

	_ [0x10 - 0xe]m.U32

	config_0 [mmu_n_rqe_queue]mmu_pipe_reg32

	_ [0x20 - 0x1b]m.U32

	config_1 [mmu_n_rqe_queue]mmu_pipe_reg32

	_ [0x30 - 0x2b]m.U32

	min_guaranteed_cells [mmu_n_rqe_queue]mmu_pipe_reg32

	_ [0x40 - 0x3b]m.U32

	color_shared_limits [mmu_n_rqe_queue]mmu_pipe_reg32

	_ [0x50 - 0x4b]m.U32

	color_reset_offset [mmu_n_rqe_queue]mmu_pipe_reg32

	_ [0x60 - 0x5b]m.U32

	min_guaranteed_cells_used [mmu_n_rqe_queue]mmu_pipe_xreg32

	_ [0x70 - 0x6b]m.U32

	shared_cells_used [mmu_n_rqe_queue]mmu_pipe_xreg32

	_ [0x80 - 0x7b]m.U32

	calculated_color_resume_limits [mmu_n_rqe_queue]mmu_pipe_xreg32

	_ [0x90 - 0x8b]m.U32

	status [mmu_n_rqe_queue]mmu_pipe_xreg32

	_ [0xa0 - 0x9b]m.U32

	bst_threshold [mmu_n_rqe_queue]mmu_pipe_reg32

	_ [0xb0 - 0xab]m.U32

	bst_total_usage_counts [mmu_n_rqe_queue]mmu_pipe_xreg32

	_ [0xc0 - 0xbb]m.U32

	bst_threshold_service_pool [4]mmu_pipe_reg32

	_ [0xd0 - 0xc4]m.U32

	bst_service_pool_usage_counts [4]mmu_pipe_xreg32

	_ [0xe0 - 0xd4]m.U32

	service_pool_status [4]mmu_pipe_xreg32

	service_pool_shared_cells_used [4]mmu_pipe_xreg32

	bst_status [4]mmu_pipe_xreg32

	_ [0xf0 - 0xec]m.U32

	qe_status mmu_pipe_xreg32

	_ [0x100 - 0xf1]m.U32
}

type mmu_pipe_mems struct {
	_            [0x04040000 - 0x0]byte
	enqx_pipemem [2]m.Mem
	_            [0x08000000 - 0x040c0000]byte

	rx_admission_control struct {
		_ [4]m.Mem

		port_priority_group_config [n_rx_pipe]struct {
			entries [n_mmu_port][n_mmu_priority_group]mmu_priority_group_config_mem
			_       [mmu_per_pipe_mem_bytes - n_mmu_port*n_mmu_priority_group]m.MemElt
		}
		_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [3]m.Mem

		port_service_pool_counters [2]m.Mem

		port_service_pool_config [n_rx_pipe]struct {
			entries [n_mmu_port][n_mmu_service_pool]mmu_rx_service_pool_config_mem
			_       [mmu_per_pipe_mem_bytes - n_mmu_port*n_mmu_service_pool]m.MemElt
		}
		_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [3]m.Mem

		port_priority_group_bst [n_mmu_port][n_mmu_priority_group]m.M32
		_                       [m.MemMax - n_mmu_port*n_mmu_priority_group]m.MemElt

		port_service_pool_bst [n_mmu_port][n_mmu_service_pool]m.M32
		_                     [m.MemMax - n_mmu_port*n_mmu_service_pool]m.MemElt
	}

	_ [0x10040000 - 0x08400000]byte

	packet_header m.Mem

	port_count m.Mem

	_ [0x14000000 - 0x100c0000]byte

	copy_count m.Mem

	ccp_reseq m.Mem

	_ [0x20000000 - 0x14080000]byte

	wred struct {
		config                                m.Mem
		average_queue_size_config             m.Mem
		unicast_queue_drop_threshold          [2]m.Mem
		unicast_queue_drop_threshold_mark     m.Mem
		port_service_pool_drop_thd            m.Mem
		port_service_pool_drop_threshold_mark m.Mem
		unicast_queue_total_count             m.Mem
		unicast_queue_total_count_from_remote m.Mem
		port_service_pool_shared_count        m.Mem
	}

	_ [0x24000000 - 0x20280000]byte

	mmu_pqe_mem [2]m.Mem

	_ [0x28000000 - 0x24080000]byte

	unicast_tx_drops [n_tx_pipe]struct {
		entries [n_mmu_port][mmu_n_tx_queues]mmu_tx_counter_mem
		_       [mmu_per_pipe_mem_bytes - n_mmu_port*mmu_n_tx_queues]m.MemElt
	}
	_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

	multicast_tx_drops [n_tx_pipe]struct {
		ports    [n_mmu_data_port][mmu_n_tx_queues]mmu_tx_counter_mem
		loopback [mmu_n_tx_queues]mmu_tx_counter_mem
		cpu      [mmu_n_cpu_queues]mmu_tx_counter_mem
		_        [mmu_per_pipe_mem_bytes - (n_mmu_data_port+1)*mmu_n_tx_queues - mmu_n_cpu_queues]m.MemElt
	}
	_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

	wred_color_drops [n_mmu_pipe]struct {
		entries [mmu_per_pipe_mem_bytes]m.M64
	}
	_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

	rx_drops [n_mmu_pipe]struct {
		entries [2][n_mmu_port]mmu_rx_counter_mem
		_       [mmu_per_pipe_mem_bytes - 2*n_mmu_port]m.MemElt
	}
	_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

	_ [1]m.Mem

	wred_drops [n_tx_pipe]struct {
		data_port_entries         [n_mmu_data_port][mmu_n_tx_queues]mmu_wred_counter_mem
		cpu_loopback_port_entries [2][mmu_n_tx_cos_queues]mmu_wred_counter_mem
		_                         [mmu_per_pipe_mem_bytes - n_mmu_data_port*mmu_n_tx_queues - 2*mmu_n_tx_cos_queues]m.MemElt
	}
	_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

	_ [0x2c000000 - 0x28180000]byte

	replication_fifo_bank [2]m.Mem
	rqe_free_list         m.Mem
	rqe_link_list         m.Mem

	mmu_replication_group_initial_copy_count [3]m.Mem

	_ [0x30000000 - 0x2c1c0000]byte

	wred_drop_curve_profile [9][3]m.Mem

	_ [0x34000000 - 0x306c0000]byte

	tx_purge_queue_memory m.Mem

	_ [0x38000000 - 0x34040000]byte

	tx_admission_control struct {
		queue_to_queue_group_map [n_tx_pipe]struct {
			data_port_entries         [n_mmu_data_port][mmu_n_tx_queues]mmu_tx_queue_to_queue_group_map_mem
			cpu_loopback_port_entries [2][mmu_n_tx_cos_queues]mmu_tx_queue_to_queue_group_map_mem
			_                         [mmu_per_pipe_mem_bytes - n_mmu_data_port*mmu_n_tx_queues - 2*mmu_n_tx_cos_queues]m.MemElt
		}
		_ [m.MemMax - n_tx_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [3]m.Mem

		queue_config [n_tx_pipe]struct {
			data_port_entries         [n_mmu_data_port][mmu_n_tx_queues]mmu_tx_queue_config_mem
			cpu_loopback_port_entries [2][mmu_n_tx_cos_queues]mmu_tx_queue_config_mem
			_                         [mmu_per_pipe_mem_bytes - n_mmu_data_port*mmu_n_tx_queues - 2*mmu_n_tx_cos_queues]m.MemElt
		}
		_ [m.MemMax - n_tx_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [2]m.Mem

		queue_offsets [n_tx_pipe]struct {
			data_port_entries         [n_mmu_data_port][mmu_n_tx_queues]mmu_tx_color_config_mem
			cpu_loopback_port_entries [2][mmu_n_tx_cos_queues]mmu_tx_color_config_mem
			_                         [mmu_per_pipe_mem_bytes - n_mmu_data_port*mmu_n_tx_queues - 2*mmu_n_tx_cos_queues]m.MemElt
		}
		_ [m.MemMax - n_tx_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [2]m.Mem

		queue_group_config [n_tx_pipe]struct {
			entries [n_mmu_port]mmu_tx_queue_config_mem
			_       [mmu_per_pipe_mem_bytes - n_mmu_port]m.MemElt
		}
		_ [m.MemMax - n_tx_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [2]m.Mem

		offset_queue_group [3]m.Mem

		queue_counters m.Mem

		queue_groups_counters m.Mem

		port_counters m.Mem

		bst_queue m.Mem

		bst_queue_group m.Mem

		bst_port m.Mem

		resume_queue m.Mem

		resume_queue_group m.Mem

		service_pool_config [n_tx_pipe]struct {
			entries [n_mmu_port][n_mmu_service_pool]mmu_tx_service_pool_config_mem
			_       [mmu_per_pipe_mem_bytes - n_mmu_port*n_mmu_service_pool]m.MemElt
		}
		_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [2]m.Mem

		resume_config [n_tx_pipe]struct {
			entries [n_mmu_port][n_mmu_service_pool]mmu_tx_color_config_mem
			_       [mmu_per_pipe_mem_bytes - n_mmu_port*n_mmu_service_pool]m.MemElt
		}
		_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [3]m.Mem
	}

	_ [0x3c000000 - 0x387c0000]byte

	multicast_admission_control struct {
		db_queue_config [n_tx_pipe]struct {
			ports    [n_mmu_data_port][mmu_n_tx_queues]m.MemElt
			loopback [mmu_n_tx_queues]m.MemElt
			cpu      [mmu_n_cpu_queues]m.MemElt
			_        [mmu_per_pipe_mem_bytes - (n_mmu_data_port+1)*mmu_n_tx_queues - mmu_n_cpu_queues]m.MemElt
		}
		_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [3]m.Mem

		db_queue_offset [n_tx_pipe]struct {
			ports    [n_mmu_data_port][mmu_n_tx_queues]m.M32
			loopback [mmu_n_tx_queues]m.M32
			cpu      [mmu_n_cpu_queues]m.M32
			_        [mmu_per_pipe_mem_bytes - (n_mmu_data_port+1)*mmu_n_tx_queues - mmu_n_cpu_queues]m.MemElt
		}
		_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [3]m.Mem

		db_queue_count m.Mem

		db_queue_bst m.Mem

		db_queue_resume m.Mem

		db_port_service_pool_bst m.Mem

		db_port_service_pool_config [n_tx_pipe]struct {
			ports [n_mmu_port][n_mmu_service_pool]mmu_multicast_db_service_pool_config_mem
			_     [mmu_per_pipe_mem_bytes - n_mmu_port*n_mmu_service_pool]m.MemElt
		}
		_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [3]m.Mem

		_ [0x40000000 - 0x3c400000]byte

		mcqe_queue_config [n_tx_pipe]struct {
			ports    [n_mmu_data_port][mmu_n_tx_queues]m.MemElt
			loopback [mmu_n_tx_queues]m.MemElt
			cpu      [mmu_n_cpu_queues]m.MemElt
			_        [mmu_per_pipe_mem_bytes - (n_mmu_data_port+1)*mmu_n_tx_queues - mmu_n_cpu_queues]m.MemElt
		}
		_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [3]m.Mem

		mcqe_queue_offset [n_tx_pipe]struct {
			ports    [n_mmu_data_port][mmu_n_tx_queues]m.M32
			loopback [mmu_n_tx_queues]m.M32
			cpu      [mmu_n_cpu_queues]m.M32
			_        [mmu_per_pipe_mem_bytes - (n_mmu_data_port+1)*mmu_n_tx_queues - mmu_n_cpu_queues]m.MemElt
		}
		_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [3]m.Mem

		mcqe_queue_count m.Mem

		mcqe_queue_bst m.Mem

		mcqe_queue_resume m.Mem

		mcqe_port_service_pool_bst m.Mem

		mcqe_port_service_pool_config [n_tx_pipe]struct {
			ports [n_mmu_port][n_mmu_service_pool]mmu_multicast_mcqe_service_pool_config_mem
			_     [mmu_per_pipe_mem_bytes - n_mmu_port*n_mmu_service_pool]m.MemElt
		}
		_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

		_ [3]m.Mem
	}
}

type mmu_sc_reg32 [1 << m.Log2NPorts]m.Gu32

func (r *mmu_sc_reg32) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint32) {
	(*m.Gu32)(&r[0]).Get(&q.DmaRequest, a, BlockMmuSc, c, v)
}
func (r *mmu_sc_reg32) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint32) {
	(*m.Gu32)(&r[0]).Set(&q.DmaRequest, a, BlockMmuSc, c, v)
}

type mmu_sc_reg64 [1 << m.Log2NPorts]m.Gu64

func (r *mmu_sc_reg64) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint64) {
	(*m.Gu64)(&r[0]).Get(&q.DmaRequest, a, BlockMmuSc, c, v)
}
func (r *mmu_sc_reg64) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint64) {
	(*m.Gu64)(&r[0]).Set(&q.DmaRequest, a, BlockMmuSc, c, v)
}

type mmu_sc_gpreg32 m.Pu32
type mmu_sc_xreg32 [1 << m.Log2NPorts]mmu_sc_gpreg32

func (r *mmu_sc_gpreg32) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint32) {
	(*m.Pu32)(r).Get(&q.DmaRequest, a, BlockMmuSc, c, v)
}
func (r *mmu_sc_gpreg32) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint32) {
	(*m.Pu32)(r).Set(&q.DmaRequest, a, BlockMmuSc, c, v)
}

type mmu_sc_gpreg64 m.Pu64
type mmu_sc_xreg64 [1 << m.Log2NPorts]mmu_sc_gpreg64

func (r *mmu_sc_gpreg64) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint64) {
	(*m.Pu64)(r).Get(&q.DmaRequest, a, BlockMmuSc, c, v)
}
func (r *mmu_sc_gpreg64) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint64) {
	(*m.Pu64)(r).Set(&q.DmaRequest, a, BlockMmuSc, c, v)
}

type mmu_sc_preg32 m.Pu32
type mmu_sc_portreg32 [1 << m.Log2NPorts]mmu_sc_preg32

func (r *mmu_sc_preg32) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint32) {
	(*m.Pu32)(r).Get(&q.DmaRequest, a, BlockMmuSc, c, v)
}
func (r *mmu_sc_preg32) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint32) {
	(*m.Pu32)(r).Set(&q.DmaRequest, a, BlockMmuSc, c, v)
}

type mmu_slice_controller struct {
	_ [0x04000100 - 0x0]byte

	asf_eport_config mmu_sc_portreg32

	_ [0x45 - 0x2]m.U32

	asf_iport_config mmu_sc_portreg32

	enqs struct {
		_ [0xce - 0x46]m.U32

		asf_error mmu_sc_xreg32

		control mmu_sc_xreg32

		port_non_empty_bitmap [2]mmu_sc_xreg64

		_ [0xd4 - 0xd2]m.U32

		debug mmu_sc_xreg32

		_ [0xd6 - 0xd5]m.U32

		no_destination_drops mmu_sc_xreg64

		_ [0x08000200 - 0x0400d700]byte
	}

	toq struct {
		fatal_error                 mmu_sc_xreg32
		multicast_cache_debug       mmu_sc_xreg32
		unicast_cache_debug         mmu_sc_xreg32
		multicast_cache_count_debug mmu_sc_xreg32
		unicast_cache_count_debug   mmu_sc_xreg32
		_                           [0x28 - 0x07]m.U32
		init                        mmu_sc_xreg32
		debug                       mmu_sc_xreg32
		_                           [0x0c100000 - 0x08002a00]byte
	}

	l3_multicast_port_aggregate_id mmu_sc_portreg32

	_ [0x10000000 - 0x0c100100]byte

	cfap struct {
		config mmu_sc_xreg32

		init mmu_sc_xreg32

		enter_full_threshold mmu_sc_xreg32

		exit_full_threshold mmu_sc_xreg32

		read_pointer mmu_sc_xreg32

		_ [0x10 - 0x05]m.U32

		bank_full_limit [mmu_n_banks]mmu_sc_xreg32

		_ [0x30 - 0x1f]m.U32

		bank_status [mmu_n_banks]mmu_sc_xreg32

		_ [0x70 - 0x3f]m.U32

		debug         mmu_sc_xreg32
		debug_scratch [3]mmu_sc_xreg32

		bst_status mmu_sc_xreg32

		bst_threshold mmu_sc_xreg32

		arbiter_mask mmu_sc_xreg32

		_ [0x80 - 0x77]m.U32

		ecc_multi_bit_enable mmu_sc_xreg32

		ecc_single_bit_enable mmu_sc_xreg32
	}

	_ [0x14000000 - 0x10008200]byte

	mtro_refresh_config mmu_sc_xreg32

	_ [1]m.U32

	mtro_port_entity_disable mmu_sc_xreg64

	_ [0x28000000 - 0x14000300]byte

	prio2cos_profile [4][16]mmu_sc_reg32

	xport_to_mmu_bkp mmu_sc_portreg32

	_ [0x50 - 0x41]m.U32

	port_llfc_config mmu_sc_portreg32

	_ [0x34000100 - 0x28005100]byte

	queue_scheduler struct {
		port_flush mmu_sc_xreg64

		port_empty_status mmu_sc_xreg64

		rqe_snapshot mmu_sc_xreg32

		dd struct {
			config [2]mmu_sc_reg32

			timer_enable [2]mmu_sc_xreg64

			timer [2]mmu_sc_portreg32

			timer_status [2]mmu_sc_xreg64

			timer_status_mask [2]mmu_sc_xreg64

			port_config mmu_sc_portreg32

			_ [0x15 - 0x0f]m.U32
		}

		port_config mmu_sc_portreg32

		l0_config mmu_sc_portreg32

		l1_unicast_config   mmu_sc_portreg32
		l1_multicast_config mmu_sc_portreg32

		cpu_port_config                   mmu_sc_reg32
		cpu_l0_config                     mmu_sc_reg32
		cpu_l1_multicast_queue_mask       mmu_sc_reg64
		cpu_l1_multicast_queue_config     mmu_sc_reg64
		cpu_l1_multicast_queue_l0_mapping [48]mmu_sc_reg32
	}

	_ [0x38000000 - 0x34004d00]byte

	mmu_port_credit mmu_sc_portreg32

	_ [0x40 - 0x01]m.U32

	asf_credit_threshold_hi mmu_sc_portreg32

	_ [0x80 - 0x41]m.U32

	mmu_1dbg_c mmu_sc_xreg32

	_ [1]m.U32

	mmu_dbg_c [3]mmu_sc_xreg32

	_ [0x90 - 0x85]m.U32

	mmu_1dbg_a mmu_sc_xreg32

	_ [0x38040000 - 0x38009100]byte

	tdm tdm_regs

	misc_config mmu_sc_reg32

	_ [1]m.U32

	interrupt_enable mmu_sc_reg32
	interrupt_status mmu_sc_xreg32
	interrupt_clear  mmu_sc_xreg32

	_ [0x7 - 0x5]m.U32

	toq_multicast_config_0 mmu_sc_reg32

	toq_multicast_config_1 mmu_sc_xreg32

	toq_multicast_config_2 mmu_sc_xreg32

	start_by_start_error mmu_sc_xreg64
}

type mmu_slice_mems struct {
	_ [0x08040000 - 0x0]byte

	cell_link m.Mem
	pkt_link  m.Mem
	mcqe      m.Mem
	mcqn      m.Mem
	mcfp      m.Mem
	ucqdb_x   m.Mem
	mcqdb_x   [2]m.Mem
	ucqdb_y   m.Mem
	mcqdb_y   [2]m.Mem
	pdb       [2]m.Mem

	_ [0x0c040000 - 0x08380000]byte

	replication struct {
		state      m.Mem
		group_info m.Mem
		head       m.Mem
		list       m.Mem

		_ [0x10040000 - 0x0c140000]byte
	}

	cfap_bank [mmu_n_banks]m.Mem

	_ [0x14000000 - 0x10400000]byte

	tx_metering struct {
		config    [3]m.Mem
		bucket    m.Mem
		l1        m.Mem
		l1_bucket m.Mem
		l0        m.Mem
		l0_bucket m.Mem

		_ [0x34000000 - 0x14200000]byte
	}

	queue_scheduler struct {
		l2_accumulated_compensation m.Mem
		l2_credit                   m.Mem
		l1_accumulated_compensation m.Mem
		l1_credit                   m.Mem
		l1_weight                   [n_tx_pipe]struct {
			unicast                  [n_mmu_data_port][mmu_n_tx_queues]m.M32
			unicast_cpu_management   [mmu_n_tx_queues]m.M32
			multicast                [n_mmu_data_port][mmu_n_tx_queues]m.M32
			unicast_loopback         [mmu_n_tx_queues]m.M32
			multicast_cpu_management [mmu_n_tx_queues]m.M32

			_ [mmu_per_pipe_mem_bytes - (2*n_mmu_data_port+3)*mmu_n_tx_queues]m.MemElt
		}
		_                           [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt
		l0_accumulated_compensation m.Mem
		l0_credit                   m.Mem
		l0_weight                   [n_tx_pipe]struct {
			unicast [n_mmu_port][mmu_n_tx_queues]m.M32

			_ [mmu_per_pipe_mem_bytes - n_mmu_port*mmu_n_tx_queues]m.MemElt
		}
		_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt
	}

	_ [0x38000000 - 0x34200000]byte

	tdm_calendar [2]struct {
		entries [n_tx_pipe][mmu_per_pipe_mem_bytes]tdm_calendar_mem
		_       [m.MemMax - n_tx_pipe*mmu_per_pipe_mem_bytes]m.MemElt
	}

	_ [0x50000000 - 0x38080000]byte

	cbp_data_slices [2]struct {
		entries [2][16][m.MemMax / mmu_per_pipe_mem_bytes][mmu_per_pipe_mem_bytes]mmu_cell_buffer_pool_mem

		_ [0x54000000 - 0x50800000]byte
	}
}
