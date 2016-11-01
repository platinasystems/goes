// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/sbus"
)

const (
	// See Reference for definitions of xpe, slice and layer.
	n_mmu_xpe               = 4
	n_mmu_slice_controllers = 2
	n_mmu_layer             = 2
	mmu_per_pipe_mem_bytes  = 0x8000
	mmu_n_cells_per_bank    = 1536
	mmu_n_banks             = 15
	mmu_n_cells_per_xpe     = mmu_n_banks * mmu_n_cells_per_bank
	// Each packet has 64 bytes of state stored in MMU.
	// This metadata originates from rx pipe lookup, parser, etc.
	mmu_per_packet_metadata_overhead = 64
	// MMU chops packets into fixed-sized cells of 208 bytes.
	mmu_cell_bytes = 208
	// Reserve CFAP Cells per XPE NB.720
	mmu_reserved_cfap_cells = 0
	// maximum packet bytes
	mmu_max_packet_bytes = 9416
	// Number of tx queues: 8 cos + SC system control + QM queue management (?)
	mmu_n_tx_cos_queues = 8
	mmu_n_tx_queues     = mmu_n_tx_cos_queues + 2
	mmu_n_cpu_queues    = 48
	// 8 cos queues plus cpu lo & hi plus mirror
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
	// Index is mmu port number. 0-255
	mmuBaseTypeRxPort sbus.Address = iota << 23
	mmuBaseTypeTxPort
	// Index is rx_pipe/tx_pipe 0-3
	mmuBaseTypeRxPipe
	mmuBaseTypeTxPipe
	// Index is always 0.
	mmuBaseTypeChip
	// Index is XPE (0-3)
	mmuBaseTypeXpe
	// Index is Slice (0-1)
	mmuBaseTypeSlice
	// Index is Layer (0-1)
	mmuBaseTypeLayer
)

type mmu_global_reg32 m.Reg32

func (r *mmu_global_reg32) geta(q *DmaRequest, c sbus.AccessType, v *uint32) {
	(*m.Reg32)(r)[0].Get(&q.DmaRequest, mmuBaseTypeChip, BlockMmuGlobal, c, v)
}
func (r *mmu_global_reg32) seta(q *DmaRequest, c sbus.AccessType, v uint32) {
	(*m.Reg32)(r)[0].Set(&q.DmaRequest, mmuBaseTypeChip, BlockMmuGlobal, c, v)
}

func (r *mmu_global_reg32) get(q *DmaRequest, v *uint32) { r.geta(q, sbus.Single, v) }
func (r *mmu_global_reg32) set(q *DmaRequest, v uint32)  { r.seta(q, sbus.Single, v) }

func (r *mmu_global_reg32) getDo(q *DmaRequest) (v uint32) {
	r.get(q, &v)
	q.Do()
	return v
}

type mmu_global_reg64 m.Reg64

func (r *mmu_global_reg64) geta(q *DmaRequest, c sbus.AccessType, v *uint64) {
	(*m.Reg64)(r).Get(&q.DmaRequest, mmuBaseTypeChip, BlockMmuGlobal, c, v)
}
func (r *mmu_global_reg64) seta(q *DmaRequest, c sbus.AccessType, v uint64) {
	(*m.Reg64)(r).Set(&q.DmaRequest, mmuBaseTypeChip, BlockMmuGlobal, c, v)
}

func (r *mmu_global_reg64) get(q *DmaRequest, v *uint64) { r.geta(q, sbus.Single, v) }
func (r *mmu_global_reg64) set(q *DmaRequest, v uint64)  { r.seta(q, sbus.Single, v) }

type mmu_global_preg32 m.Preg32
type mmu_global_portreg32 [1 << m.Log2NRegPorts]mmu_global_preg32

func (r *mmu_global_preg32) geta(q *DmaRequest, c sbus.AccessType, v *uint32) {
	(*m.Preg32)(r).Get(&q.DmaRequest, mmuBaseTypeRxPort, BlockMmuGlobal, c, v)
}
func (r *mmu_global_preg32) seta(q *DmaRequest, c sbus.AccessType, v uint32) {
	(*m.Preg32)(r).Set(&q.DmaRequest, mmuBaseTypeRxPort, BlockMmuGlobal, c, v)
}
func (r *mmu_global_preg32) set(q *DmaRequest, v uint32) { r.seta(q, sbus.Single, v) }

type mmu_global_regs struct {
	_ [0x08000000 - 0x0]byte

	// [4] Override sbus splitter error check
	// [3] bst mode: 1 => max usage count, 0 => current usage count
	// [2] hardware parity/ecc enable
	// [1] start memory initialization
	// [0] enable refresh clock
	misc_config mmu_global_reg32

	_ [1]m.Reg32

	// [0] enable
	parity_interrupt struct {
		enable                  mmu_global_reg32
		status                  mmu_global_reg32
		status_write_1_to_clear mmu_global_reg32
	}

	// [6:5] slice controller interrupt blocks S,R
	// [4:1] xpe interrupt SA SB RA RB
	// [0] global interrupt
	interrupt_enable mmu_global_reg32

	// Read only; bits as above.  Clear interrupt at source.
	interrupt_status mmu_global_reg32

	_ [0xb - 0x7]m.Reg32

	// [11:8] bst enable tx admission control XPE RA SA RB SB
	// [7:4]  bst enable rx admission control XPE RA SA RB SB
	// [3:0]  bsd enable CFAP XPE RA SA RB SB
	bst_tracking_enable mmu_global_reg32

	// [2] tx admission control
	// [1] rx admission control
	// [0] CFAP
	bst_hw_snapshot_reset mmu_global_reg32

	// Bits same as for bst_tracking_enable.
	bst_hw_snapshot_enable mmu_global_reg32

	_ [0x1000 - 0xe]m.Reg32

	// Default value: 0xff
	device_port_by_mmu_port mmu_global_portreg32

	_ [0x1100 - 0x1001]m.Reg32

	physical_port_by_mmu_port mmu_global_portreg32

	_ [0x1200 - 0x1101]m.Reg32

	// Default value: 0xff
	global_physical_port_by_mmu_port mmu_global_portreg32
}

type mmu_global_mems struct{}

type mmu_xpe_reg32 [1 << m.Log2NRegPorts]m.Greg32

func (r *mmu_xpe_reg32) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint32) {
	(*m.Greg32)(&r[0]).Get(&q.DmaRequest, a, BlockMmuXpe, c, v)
}
func (r *mmu_xpe_reg32) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint32) {
	(*m.Greg32)(&r[0]).Set(&q.DmaRequest, a, BlockMmuXpe, c, v)
}

type mmu_xpe_gpreg32 m.Preg32
type mmu_xpe_xreg32 [1 << m.Log2NRegPorts]mmu_xpe_gpreg32

func (r *mmu_xpe_gpreg32) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint32) {
	(*m.Preg32)(r).Get(&q.DmaRequest, a, BlockMmuXpe, c, v)
}
func (r *mmu_xpe_gpreg32) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint32) {
	(*m.Preg32)(r).Set(&q.DmaRequest, a, BlockMmuXpe, c, v)
}

type mmu_xpe_gpreg64 m.Preg64
type mmu_xpe_xreg64 [1 << m.Log2NRegPorts]mmu_xpe_gpreg64

func (r *mmu_xpe_gpreg64) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint64) {
	(*m.Preg64)(r).Get(&q.DmaRequest, a, BlockMmuXpe, c, v)
}
func (r *mmu_xpe_gpreg64) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint64) {
	(*m.Preg64)(r).Set(&q.DmaRequest, a, BlockMmuXpe, c, v)
}

type mmu_xpe_preg32 m.Preg32
type mmu_xpe_portreg32 [1 << m.Log2NRegPorts]mmu_xpe_preg32

func (r *mmu_xpe_preg32) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint32) {
	(*m.Preg32)(r).Get(&q.DmaRequest, a, BlockMmuXpe, c, v)
}
func (r *mmu_xpe_preg32) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint32) {
	(*m.Preg32)(r).Set(&q.DmaRequest, a, BlockMmuXpe, c, v)
}

type mmu_xpe_regs struct {
	_ [0x08000000 - 0x0]byte

	// Rx admission control thresholds.
	rx_admission_control struct {
		// [17] rx enable.  when 0, traffic in flight will go out.
		// [16] allows port to generate pause frames
		// [15:0] per-priority xon enable
		port_rx_and_pause_enable mmu_xpe_portreg32

		_ [0x7 - 0x1]m.Reg32

		// [14:0]
		global_headroom_limit mmu_xpe_xreg32

		_ [0x22 - 0x8]m.Reg32

		// Largest mtu for any port making use of global headroom.
		max_packet_cells mmu_xpe_reg32

		_ [0x70 - 0x23]m.Reg32

		// 3 bit PG by 4 bit input packet priority.  8 PGs per register: 0 => packet pri 0-7, 1 => packet pri 8-15
		port_priority_group [2]mmu_xpe_portreg32

		_ [0x73 - 0x72]m.Reg32

		// 2 bit service pool for 8 priority groups.
		port_service_pool_for_priority_group mmu_xpe_portreg32

		// 2 bit headroom pool for 8 priority groups.
		port_headroom_pool_for_priority_group mmu_xpe_portreg32

		_ [0x7b - 0x75]m.Reg32

		// read-only status.
		global_headroom_cells_in_use          mmu_xpe_xreg32
		global_headroom_reserved_cells_in_use mmu_xpe_xreg32
		_                                     [0x100 - 0x7d]m.Reg32
		service_pool_n_shared_cells_in_use    [n_mmu_service_pool_plus_public]mmu_xpe_xreg32
		_                                     [0x10a - 0x105]m.Reg32

		service_pool_shared_cell_limit       [n_mmu_service_pool_plus_public]mmu_xpe_xreg32
		_                                    [0x114 - 0x10f]m.Reg32
		service_pool_cell_reset_limit_offset [n_mmu_service_pool]mmu_xpe_xreg32
		_                                    [0x11c - 0x118]m.Reg32
		cell_spap_yellow_offset              [n_mmu_service_pool]mmu_xpe_xreg32
		cell_spap_red_offset                 [n_mmu_service_pool]mmu_xpe_xreg32

		_ [0x131 - 0x124]m.Reg32

		// [7:4] public pool enable
		// [3:0] color enable (0 => all colors are the same).
		service_pool_config mmu_xpe_xreg32

		_ [0x134 - 0x132]m.Reg32

		// [11:8] service pool using all shared cells.
		//        cleared when shared cell count drops below resume limit.
		// [7:0] as above but for priority groups.
		port_limit_states mmu_xpe_portreg32

		_ [0x138 - 0x135]m.Reg32

		// [7:0] priority group is in xoff state.
		priority_group_xoff_status mmu_xpe_portreg32

		// 3 color bits (green, yellow, red) x 4 service pools; 1 => pool is dropping packets.
		pool_drop_state mmu_xpe_xreg32

		_ [0x140 - 0x13a]m.Reg32

		// [3:0] one bit foreach headroom pool
		headroom_pool_peak_count_update_enable mmu_xpe_xreg32

		// [3:0] headroom pool limit has been reached.
		headroom_pool_status mmu_xpe_xreg32

		_ [0x150 - 0x142]m.Reg32

		headroom_pool_n_cells_in_use [4]mmu_xpe_xreg32

		// [14:0] max number of cells for all ports belonging to headroom pool.
		headroom_pool_cell_limit [4]mmu_xpe_xreg32

		headroom_pool_peak_cell_count [4]mmu_xpe_xreg32

		_ [0x170 - 0x15c]m.Reg32

		buffer_stats_tracking struct {
			// Stats event triggers when cell count crosses one of 8 threshold levels.
			priority_group_shared_cell_threshold   [8]mmu_xpe_reg32
			priority_group_headroom_cell_threshold [8]mmu_xpe_reg32
			service_pool_shared_cell_threshold     [8]mmu_xpe_reg32
			_                                      [0x190 - 0x188]m.Reg32

			// read-only shared usage counts; clear by writing corresponding bit in misc control.
			service_pool_global_shared_cells [n_mmu_service_pool_plus_public]mmu_xpe_reg32
			_                                [8 - n_mmu_service_pool_plus_public]mmu_xpe_reg32

			service_pool_global_shared_cell_threshold [8]mmu_xpe_xreg32

			// [3] priority group shared threshold crossed
			// [2] priority group headroom threshold crossed
			// [1] service pool ...
			// [0] global pool ...
			trigger_status_type mmu_xpe_xreg32

			// [28:23] port that triggered priority group threshold crossing.
			// [22:20] priority group that triggered.
			// [19:14] port that triggered priority group headroom threshold crossing.
			// [13:11] priority group ...
			// [10:5] port ...
			// [4:3] service pool ...
			// [2:0] service pool that triggered global pool shared threshold crossing.
			trigger_status mmu_xpe_xreg32
		}
	}

	_ [0x10002b00 - 0x0801a200]byte

	cut_through_purge_count mmu_xpe_xreg32

	_ [0x14001000 - 0x10002c00]byte

	ccp_status mmu_xpe_xreg32

	_ [0x20000000 - 0x14001100]byte

	time_domain [4]mmu_xpe_reg32

	_ [0x6 - 0x4]m.Reg32

	wred_refresh_control mmu_xpe_reg32

	wred_congestion_notification_resolution_table [4]mmu_xpe_reg32

	wred_pool_congestion_limit [4]mmu_xpe_reg32

	_ [0x24000000 - 0x20000f00]byte

	pqe [2]struct {
		fifo_empty_port_bitmap [2]mmu_xpe_xreg32

		fifo_overflow_port_bitmap [2]mmu_xpe_xreg32

		_ [0x5 - 0x4]m.Reg32

		fifo_pointer_equal_bitmap [2]mmu_xpe_xreg32

		_ [0x8 - 0x7]m.Reg32

		qcn_fifo_empty_port_bitmap [2]mmu_xpe_xreg32

		_ [0x10 - 0xa]m.Reg32
	}

	_ [0x28000000 - 0x24002000]byte

	// [0] resets all non-memory counters (rqe drops, cbp drops, tx packet counter)
	clear_counters mmu_xpe_reg32

	_ [0x10 - 0x1]m.Reg32

	// [31:30] color
	// [29:18] multicast COS, priority value, or queue.
	// [17:11] destination port.
	// [10:4] source port.
	// 7 bit ports => xpe0: 0-33 pipe 0 ports; 34-67 pipe 1 ports, etc.
	tx_counter_config [8]mmu_xpe_xreg32

	_ [0x20 - 0x18]m.Reg32

	// [35:0]
	tx_counter_packets [8]mmu_xpe_xreg64

	_ [0x110 - 0x28]m.Reg32

	// [35:0]
	multicast_replication_fifo_drop_packets [mmu_n_rqe_queue]mmu_xpe_xreg64
	_                                       [0x120 - 0x11b]m.Reg32

	// [43:0]
	multicast_replication_fifo_drop_bytes [mmu_n_rqe_queue]mmu_xpe_xreg64
	_                                     [0x12f - 0x12b]m.Reg32

	// [39:0]
	data_buffer_full_packets_dropped mmu_xpe_xreg64

	_ [0x200 - 0x130]m.Reg32

	// [35:0]
	multicast_replication_fifo_red_drop_packets [mmu_n_rqe_queue]mmu_xpe_xreg64

	_ [0x220 - 0x20b]m.Reg32

	// [35:0]
	multicast_replication_fifo_yellow_drop_packets [mmu_n_rqe_queue]mmu_xpe_xreg64

	_ [0x2c003600 - 0x28022b00]byte

	// [0] 1 => strict priority, 0 => WERR
	rqe_priority_scheduling_type [mmu_n_rqe_queue]mmu_xpe_reg32

	_ [0x60 - 0x41]m.Reg32

	// [6:0] Weights for each RQE priority queue.
	rqe_priority_werr_weight [mmu_n_rqe_queue]mmu_xpe_reg32

	_ [0x88 - 0x6b]m.Reg32

	// [33:0] port bitmap 0 => cos mode 1, 1 => cos mode 2.
	port_cos_mode mmu_xpe_xreg64

	_ [0x110 - 0x89]m.Reg32

	port_initial_copy_count_width mmu_xpe_portreg32

	_ [0x38001000 - 0x2c011100]byte

	// Tx admission control thresholds.
	tx_admission_control struct {
		// [0] enable bypass of threshold checking. For debug only.
		bypass mmu_xpe_reg32

		_ [0x12 - 0x11]m.Reg32

		config mmu_xpe_reg32

		_ [0x74 - 0x13]m.Reg32

		// [33:0] port bitmap
		tx_port_enable mmu_xpe_xreg64

		_ [0x100 - 0x75]m.Reg32

		// bitmap: 1 bit for each of 10 queues x 34 ports
		queue_drop_state_bitmap [6]mmu_xpe_xreg64

		_ [0x200 - 0x106]m.Reg32

		queue_yellow_drop_state_bitmap [6]mmu_xpe_xreg64
		_                              [0x300 - 0x206]m.Reg32

		queue_red_drop_state_bitmap [6]mmu_xpe_xreg64

		_ [0x500 - 0x306]m.Reg32

		// [33:0] one bit for each mmu port.
		queue_group_drop_state mmu_xpe_xreg64

		_ [0x520 - 0x501]m.Reg32

		queue_group_yellow_drop_state mmu_xpe_xreg64

		_ [0x540 - 0x521]m.Reg32

		queue_group_red_drop_state mmu_xpe_xreg64

		_ [0x610 - 0x541]m.Reg32

		// [33:0] port bitmap.
		per_service_pool_port_drop_state [4]mmu_xpe_xreg64

		_ [0x620 - 0x614]m.Reg32

		per_service_pool_port_yellow_drop_state [4]mmu_xpe_xreg64

		_ [0x630 - 0x624]m.Reg32

		per_service_pool_port_red_drop_state [4]mmu_xpe_xreg64

		_ [0x820 - 0x634]m.Reg32

		// [11:0] 8-cell threshold to trigger bst event for queues.
		queue_bst_threshold [8]mmu_xpe_xreg32

		// [11:0] 8-cell threshold to trigger bst event for queue groups.
		queue_group_bst_threshold [8]mmu_xpe_xreg32

		// [11:0] 8-cell bst threshold for ports.
		port_bst_threshold [8]mmu_xpe_xreg32

		_ [0x850 - 0x838]m.Reg32

		bst_status mmu_xpe_xreg32

		_ [0x900 - 0x851]m.Reg32

		port_e2ecc_cos_spid mmu_xpe_portreg32

		_ [0xa0b - 0x901]m.Reg32

		// [13:8] port
		// [7:0] cos bitmap, 1 => reset congestion state for port, COS
		congestion_state_reset mmu_xpe_xreg32
	}

	_ [0x3c000400 - 0x380a0c00]byte

	// Multicast admission control (thresholding).
	// DB = data buffer; QE = MC queue element (a.k.a. mcqe)
	multicast_admission_control struct {
		db struct {
			// [14:0] shared limit not including guaranteed min cells.
			service_pool_shared_limit [4]mmu_xpe_reg32

			// [11:0] 8-cell color shared limit
			service_pool_yellow_shared_limit [4]mmu_xpe_reg32
			service_pool_red_shared_limit    [4]mmu_xpe_reg32

			_ [0x12 - 0x10]m.Reg32

			config mmu_xpe_reg32

			_ [0x30 - 0x13]m.Reg32

			// [2:0] per-port bst threshold select.
			service_pool_bst_threshold_profile_select [4]mmu_xpe_portreg32

			_ [0x40 - 0x34]m.Reg32

			// [14:0] cell count read-only
			service_pool_shared_count [4]mmu_xpe_xreg32

			_ [0x4c - 0x44]m.Reg32

			// [14:0] as above but multicast.
			service_pool_multicast_shared_count [4]mmu_xpe_xreg32

			// 14:0 COUNT RO The number of "shared" cells currently used by an output port across all of its COS queues.
			port_service_pool_shared_count [4]mmu_xpe_portreg32

			_ [0x6c - 0x54]m.Reg32

			// per color drop states.
			// [11:8] red
			// [7:4] yellow
			// [3:0] green
			service_pool_drop_states mmu_xpe_xreg32

			_ [0x74 - 0x6d]m.Reg32

			// [33:0] port bitmap
			port_tx_enable mmu_xpe_xreg64

			_ [0x80 - 0x75]m.Reg32

			// [33:0] port bitmap
			port_service_pool_yellow_drop_state [4]mmu_xpe_xreg64

			// [11:0] 8-cell resume limit
			service_pool_resume_limit [n_packet_color][4]mmu_xpe_reg32

			_ [0xa0 - 0x90]m.Reg32

			port_service_pool_red_drop_state [4]mmu_xpe_xreg64

			_ [0x260 - 0xa4]m.Reg32

			port_service_pool_drop_state [4]mmu_xpe_xreg64

			_ [0x500 - 0x264]m.Reg32

			// [14:0] bst thresholds for 48 cpu queues.
			cpu_queue_bst_threshold_profile [8]mmu_xpe_xreg32

			_ [0x540 - 0x508]m.Reg32

			// [11:0] 8-cell resume limit.
			queue_resume_offset_profile_yellow [8]mmu_xpe_reg32

			// [11:0] 8-cell resume limit.
			queue_resume_offset_profile_red [8]mmu_xpe_reg32

			queue_e2e_ds mmu_xpe_portreg32

			_ [0x638 - 0x551]m.Reg32

			service_pool_mcuc_bst_threshold [4]mmu_xpe_reg32

			service_pool_mcuc_bst_stat [4]mmu_xpe_xreg32

			queue_multicast_bst_threshold_profile [8]mmu_xpe_reg32

			_ [0x650 - 0x648]m.Reg32

			device_bst_status mmu_xpe_xreg64

			_ [0x6a0 - 0x651]m.Reg32

			service_pool_multicast_bst_threshold [4]mmu_xpe_reg32

			service_pool_multicast_bst_status [4]mmu_xpe_xreg32

			_ [0x6b0 - 0x6a8]m.Reg32

			port_service_pool_bst_threshold [8]mmu_xpe_reg32

			_ [0x700 - 0x6b8]m.Reg32

			queue_e2e_spid mmu_xpe_portreg32

			_ [0x950 - 0x701]m.Reg32

			enable_eccp_mem mmu_xpe_reg32

			enable_correctable_error_reporting mmu_xpe_reg32

			_ [0xa00 - 0x952]m.Reg32

			queue_e2e_ds_en mmu_xpe_portreg32

			_ [0x40000400 - 0x3c0a0100]byte
		}

		mcqe struct {
			// [10:0] units of 4 mcqe entries
			pool_shared_limit [4]mmu_xpe_reg32

			// [9:0] units of 8 mcqe entries
			pool_yellow_shared_limit [4]mmu_xpe_reg32
			pool_red_shared_limit    [4]mmu_xpe_reg32

			_ [0x12 - 0x10]m.Reg32

			// 12:9 port service pool color enable
			// 8:5 global pool color enable
			// 4:1 pool congestion status enable
			// 0:0 early e2e drop status
			config mmu_xpe_reg32

			_ [0x30 - 0x13]m.Reg32

			// [2:0]
			port_service_pool_bst_threshold_profile_select [4]mmu_xpe_xreg32

			_ [0x40 - 0x34]m.Reg32

			// [12:0] shared mcqe entries
			pool_shared_count [4]mmu_xpe_xreg32

			_ [0x50 - 0x44]m.Reg32

			// [12:0] mcqe entries
			port_service_pool_shared_count [4]mmu_xpe_portreg32

			_ [0x6c - 0x54]m.Reg32

			// 11:8 red
			// 7:4 yellow
			// 3:0 green service pools
			pool_drop_states mmu_xpe_xreg32

			_ [0x74 - 0x6d]m.Reg32

			// [33:0] tx port enable
			port_tx_enable mmu_xpe_xreg64

			_ [0x80 - 0x75]m.Reg32

			port_service_pool_yellow_drop_state [4]mmu_xpe_xreg64

			// 9:0 units of 8 mcqe entries
			pool_resume_limit [n_packet_color][4]mmu_xpe_reg32

			_ [0xa0 - 0x90]m.Reg32

			// [33:0] port bitmap
			port_service_pool_red_drop_state [4]mmu_xpe_xreg64
			_                                [0x260 - 0xa4]m.Reg32
			port_service_pool_drop_state     [4]mmu_xpe_xreg64
			_                                [0x500 - 0x264]m.Reg32

			// [12:0] 0xfff
			cpu_queue_bst_threshold_profile [8]mmu_xpe_xreg32

			_ [0x540 - 0x508]m.Reg32

			// [9:0] units of 8 mcqe
			queue_resume_offset_profile_yellow [8]mmu_xpe_reg32
			queue_resume_offset_profile_red    [8]mmu_xpe_reg32

			_ [0x640 - 0x550]m.Reg32

			queue_bst_threshold_profile [8]mmu_xpe_reg32

			_ [0x650 - 0x648]m.Reg32

			bst_status mmu_xpe_xreg32

			_ [0x6a0 - 0x651]m.Reg32

			pool_multicast_bst_threshold [4]mmu_xpe_reg32

			pool_multicast_bst_status [4]mmu_xpe_xreg32

			_ [0x6b0 - 0x6a8]m.Reg32

			port_service_pool_bst_threshold [8]mmu_xpe_reg32

			_ [0x850 - 0x6b8]m.Reg32

			// [6] port service pool bst
			// [5] queue bst 0x1
			// [4] port service pool config 0x1
			// [3] queue resume 0x1
			// [2] queue count 0x1
			// [1] queue offset 0x1
			// [0] queue config 0x1
			// Default: 0x7f all enabled
			enable_ecc_parity mmu_xpe_reg32

			// 10 port service pool config c
			// 9  port service pool config b
			// 8  port service pool config a
			// 7 queue resume
			// 6 queue count
			// 5 queue offset c
			// 4 queue offset b
			// 3 queue offset a
			// 2 queue config c
			// 1 queue config b
			// 0 queue config a
			enable_ecc_error_reporting mmu_xpe_reg32
		}
	}

	_ [0x44000000 - 0x40085200]byte

	db mmu_xpe_rqe_admission_control_regs

	_ [0x48010000 - 0x44010000]byte

	qe mmu_xpe_rqe_admission_control_regs

	_ [0x4c000100 - 0x48020000]byte

	// 4 deq1 not ip error
	// 3 deq0 not ip error
	// 2 bst tx threshold event triggered
	// 1 bst rx threshold event triggered
	// 0 memory parity error
	interrupt_enable mmu_xpe_reg32
	interrupt_status mmu_xpe_xreg32
	intterupt_clear  mmu_xpe_xreg32
}

type mmu_xpe_rqe_admission_control_regs struct {
	_ [1]m.Reg32

	// [2] clear drop state when queue/port service pool shared limit/color limit is written
	// [1] mop policy 1b "ticket policy"
	// [0] mop policy 1a
	config mmu_xpe_reg32

	// [26:15] 8-cell unit resume limit
	// [14:0] 1-cell unit shared limit
	service_pool_config [4]mmu_xpe_reg32

	// [23:12] 8-cell shared yellow limit
	// [11:0] same for red cells
	service_pool_per_color_shared_limit [4]mmu_xpe_reg32

	// [23:12] 8-cell yello resume limit
	// [11:0] same for red cells
	resume_color_limit [4]mmu_xpe_reg32

	_ [0x10 - 0xe]m.Reg32

	// [5:4] service pool for this queue
	// [3] color threshold percentage of shared limit
	// [2] limit enable (0 => never drop)
	// [1] dynamic treshold
	// [0] color enable
	config_0 [mmu_n_rqe_queue]mmu_xpe_reg32

	_ [0x20 - 0x1b]m.Reg32

	// [26:15] 8-cell reset offset
	// [14:0] shared limit in cells or alpha if dynamic
	// In dynamic mode shared limit [3:0] : alpha = 2^(v-7) v = 0..10
	config_1 [mmu_n_rqe_queue]mmu_xpe_reg32

	_ [0x30 - 0x2b]m.Reg32

	// [14:0]
	min_guaranteed_cells [mmu_n_rqe_queue]mmu_xpe_reg32

	_ [0x40 - 0x3b]m.Reg32

	// [23:12] yellow shared limit 8-cell units
	// [11:0] same for red cells
	// In dynamic mode [2:0] represents 0 => 100%, i != 0 => 12.5% * i
	color_shared_limits [mmu_n_rqe_queue]mmu_xpe_reg32

	_ [0x50 - 0x4b]m.Reg32

	// [23:12] yellow reset offset 8-cell units
	// [11:0] same for red
	color_reset_offset [mmu_n_rqe_queue]mmu_xpe_reg32

	_ [0x60 - 0x5b]m.Reg32

	min_guaranteed_cells_used [mmu_n_rqe_queue]mmu_xpe_xreg32
	_                         [0x70 - 0x6b]m.Reg32
	shared_cells_used         [mmu_n_rqe_queue]mmu_xpe_xreg32
	_                         [0x80 - 0x7b]m.Reg32

	// [23:12] read only yellow resume limit (8-cell units)
	// [11:0] same for red
	calculated_color_resume_limits [mmu_n_rqe_queue]mmu_xpe_xreg32

	_ [0x90 - 0x8b]m.Reg32

	// [14:3] calculated resume limit (in 8-cell units)
	// [2:0] red/yellow/green drop state
	status [mmu_n_rqe_queue]mmu_xpe_xreg32

	_ [0xa0 - 0x9b]m.Reg32

	bst_threshold [mmu_n_rqe_queue]mmu_xpe_reg32

	_ [0xb0 - 0xab]m.Reg32

	bst_total_usage_counts [mmu_n_rqe_queue]mmu_xpe_xreg32

	_ [0xc0 - 0xbb]m.Reg32

	bst_threshold_service_pool [4]mmu_xpe_reg32

	_ [0xd0 - 0xc4]m.Reg32

	bst_service_pool_usage_counts [4]mmu_xpe_xreg32

	_ [0xe0 - 0xd4]m.Reg32

	// [2:0] red/yellow/green drop state for service pool
	service_pool_status [4]mmu_xpe_xreg32

	service_pool_shared_cells_used [4]mmu_xpe_xreg32

	// [7:4] triggering priority queue
	// [3:2] triggering service pool
	// [1] priority queue valid
	// [0] service pool valid
	bst_status [4]mmu_xpe_xreg32

	_ [0xf0 - 0xec]m.Reg32

	// [10] drop state
	// [9:0] the number of queue entry currently used out of the available 1k.
	qe_status mmu_xpe_xreg32

	_ [0x100 - 0xf1]m.Reg32
}

type mmu_xpe_mems struct {
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

		// [30] parity
		// [29:15] headroom cells for this port and priority group.
		// [14:0] shared cells
		// Watermark mode: max ever used; else current values.
		port_priority_group_bst [n_mmu_port][n_mmu_priority_group]m.Mem32
		_                       [m.MemMax - n_mmu_port*n_mmu_priority_group]m.MemElt

		// [15] parity
		// [14:0] number of cells used for this port and service pool.
		// Watermark mode: max ever used; else current values.
		port_service_pool_bst [n_mmu_port][n_mmu_service_pool]m.Mem32
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

	_           [0x24000000 - 0x20280000]byte
	mmu_pqe_mem [2]m.Mem
	_           [0x28000000 - 0x24080000]byte

	// [80] parity
	// [79:36] bytes dropped
	// [35:0] packets dropped
	unicast_tx_drops [n_tx_pipe]struct {
		entries [n_mmu_port][mmu_n_tx_queues]mmu_tx_counter_mem
		_       [mmu_per_pipe_mem_bytes - n_mmu_port*mmu_n_tx_queues]m.MemElt
	}
	_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

	// as above
	multicast_tx_drops [n_tx_pipe]struct {
		ports    [n_mmu_data_port][mmu_n_tx_queues]mmu_tx_counter_mem
		loopback [mmu_n_tx_queues]mmu_tx_counter_mem
		cpu      [mmu_n_cpu_queues]mmu_tx_counter_mem
		_        [mmu_per_pipe_mem_bytes - (n_mmu_data_port+1)*mmu_n_tx_queues - mmu_n_cpu_queues]m.MemElt
	}
	_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

	wred_color_drops [n_mmu_xpe]struct {
		entries [mmu_per_pipe_mem_bytes]m.Mem64
	}
	_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

	// [70] parity
	// [69:32] bytes dropped
	// [31:0] packets dropped
	// [0] => xpe 0 pipes 0 [0..33] & 3 [34..67]
	// [1] => xpe 1 pipes 0 [0..33] & 3 [34..67]
	// [2] => xpe 2 pipes 1 & 2
	// [3] => xpe 3 pipes 1 & 2
	rx_drops [n_mmu_xpe]struct {
		entries [2][n_mmu_port]mmu_rx_counter_mem
		_       [mmu_per_pipe_mem_bytes - 2*n_mmu_port]m.MemElt
	}
	_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt

	_ [1]m.Mem

	// [36] parity
	// [35:0] packets dropped
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

		// [36] parity
		// [35:30] ecc
		// [29:15] queue shared cells
		// [14:0] queue min guaranteed cells
		queue_counters m.Mem

		// as above
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

	// Multicast admission control (thresholding).
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
			ports    [n_mmu_data_port][mmu_n_tx_queues]m.Mem32
			loopback [mmu_n_tx_queues]m.Mem32
			cpu      [mmu_n_cpu_queues]m.Mem32
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
			ports    [n_mmu_data_port][mmu_n_tx_queues]m.Mem32
			loopback [mmu_n_tx_queues]m.Mem32
			cpu      [mmu_n_cpu_queues]m.Mem32
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

type mmu_sc_reg32 [1 << m.Log2NRegPorts]m.Greg32

func (r *mmu_sc_reg32) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint32) {
	(*m.Greg32)(&r[0]).Get(&q.DmaRequest, a, BlockMmuSc, c, v)
}
func (r *mmu_sc_reg32) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint32) {
	(*m.Greg32)(&r[0]).Set(&q.DmaRequest, a, BlockMmuSc, c, v)
}

type mmu_sc_reg64 [1 << m.Log2NRegPorts]m.Greg64

func (r *mmu_sc_reg64) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint64) {
	(*m.Greg64)(&r[0]).Get(&q.DmaRequest, a, BlockMmuSc, c, v)
}
func (r *mmu_sc_reg64) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint64) {
	(*m.Greg64)(&r[0]).Set(&q.DmaRequest, a, BlockMmuSc, c, v)
}

type mmu_sc_gpreg32 m.Preg32
type mmu_sc_xreg32 [1 << m.Log2NRegPorts]mmu_sc_gpreg32

func (r *mmu_sc_gpreg32) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint32) {
	(*m.Preg32)(r).Get(&q.DmaRequest, a, BlockMmuSc, c, v)
}
func (r *mmu_sc_gpreg32) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint32) {
	(*m.Preg32)(r).Set(&q.DmaRequest, a, BlockMmuSc, c, v)
}

type mmu_sc_gpreg64 m.Preg64
type mmu_sc_xreg64 [1 << m.Log2NRegPorts]mmu_sc_gpreg64

func (r *mmu_sc_gpreg64) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint64) {
	(*m.Preg64)(r).Get(&q.DmaRequest, a, BlockMmuSc, c, v)
}
func (r *mmu_sc_gpreg64) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint64) {
	(*m.Preg64)(r).Set(&q.DmaRequest, a, BlockMmuSc, c, v)
}

type mmu_sc_preg32 m.Preg32
type mmu_sc_portreg32 [1 << m.Log2NRegPorts]mmu_sc_preg32

func (r *mmu_sc_preg32) get(q *DmaRequest, a sbus.Address, c sbus.AccessType, v *uint32) {
	(*m.Preg32)(r).Get(&q.DmaRequest, a, BlockMmuSc, c, v)
}
func (r *mmu_sc_preg32) set(q *DmaRequest, a sbus.Address, c sbus.AccessType, v uint32) {
	(*m.Preg32)(r).Set(&q.DmaRequest, a, BlockMmuSc, c, v)
}

// SC = Slice Controller
type mmu_sc_regs struct {
	_ [0x04000100 - 0x0]byte

	asf_eport_config mmu_sc_portreg32

	_ [0x45 - 0x2]m.Reg32

	asf_iport_config mmu_sc_portreg32

	enqs struct {
		_ [0xce - 0x46]m.Reg32

		asf_error mmu_sc_xreg32

		// [0] clear no destination drop counter
		control mmu_sc_xreg32

		port_non_empty_bitmap [2]mmu_sc_xreg64

		_ [0xd4 - 0xd2]m.Reg32

		// 11:8 per pipe invalid source port
		// 7:4 per pipe tdm minimum spacing of 1 in 4 slots was violated
		// 3:0 per pipe SOP followed by SOP error
		debug mmu_sc_xreg32

		_ [0xd6 - 0xd5]m.Reg32

		// [39:0]
		no_destination_drops mmu_sc_xreg64

		_ [0x08000200 - 0x0400d700]byte
	}

	toq struct {
		fatal_error                 mmu_sc_xreg32
		multicast_cache_debug       mmu_sc_xreg32
		unicast_cache_debug         mmu_sc_xreg32
		multicast_cache_count_debug mmu_sc_xreg32
		unicast_cache_count_debug   mmu_sc_xreg32
		_                           [0x28 - 0x07]m.Reg32
		init                        mmu_sc_xreg32
		debug                       mmu_sc_xreg32
		_                           [0x0c100000 - 0x08002a00]byte
	}

	// [5:0]
	l3_multicast_port_aggregate_id mmu_sc_portreg32

	_ [0x10000000 - 0x0c100100]byte

	// CFAP = Cell Free Address Pool
	cfap struct {
		// 14:0 Size
		// Default value: 0x59ff = 1536 cells * 15 banks - 1
		config mmu_sc_xreg32

		// Re-Initialize CFAP Memory.
		// 14:0 bank bitmap
		init mmu_sc_xreg32

		// 14:0 threshold defining full condition 0x4dff
		enter_full_threshold mmu_sc_xreg32

		// 14:0 threshold defining exit of full condition 0x4d80
		exit_full_threshold mmu_sc_xreg32

		// [15] full status
		// [14:0] number of cell pointers that are in use.
		read_pointer mmu_sc_xreg32

		_ [0x10 - 0x05]m.Reg32

		// [9:0] 0x2ff
		bank_full_limit [mmu_n_banks]mmu_sc_xreg32

		_ [0x30 - 0x1f]m.Reg32

		// 10:10 bank is full
		// 9:0 current stack pointer in cfap bank memory
		bank_status [mmu_n_banks]mmu_sc_xreg32

		_ [0x70 - 0x3f]m.Reg32

		debug         mmu_sc_xreg32
		debug_scratch [3]mmu_sc_xreg32

		// 14:0 current count
		bst_status mmu_sc_xreg32

		// 14:0 bst trigger threshold 0x7fff
		bst_threshold mmu_sc_xreg32

		// 3:0 MASK 4 bit mask to zero out 0 <= n <= 11 LSB stackpointer bits before comparing them in the arbiter. 0x0
		arbiter_mask mmu_sc_xreg32

		_ [0x80 - 0x77]m.Reg32

		// [14:0] per bank 2 bit ecc enable
		ecc_multi_bit_enable mmu_sc_xreg32

		// [14:0] per bank 1 bit ecc enable
		ecc_single_bit_enable mmu_sc_xreg32
	}

	_ [0x14000000 - 0x10008200]byte

	mtro_refresh_config mmu_sc_xreg32

	_ [1]m.Reg32

	mtro_port_entity_disable mmu_sc_xreg64

	_ [0x28000000 - 0x14000300]byte

	prio2cos_profile [4][16]mmu_sc_reg32

	xport_to_mmu_bkp mmu_sc_portreg32

	_ [0x50 - 0x41]m.Reg32

	port_llfc_config mmu_sc_portreg32

	_ [0x34000100 - 0x28005100]byte

	queue_scheduler struct {
		// 33:0 per port bitmap
		port_flush mmu_sc_xreg64

		// 33:0 per port queue empty status
		port_empty_status mmu_sc_xreg64

		rqe_snapshot mmu_sc_xreg32

		dd struct {
			config [2]mmu_sc_reg32

			timer_enable [2]mmu_sc_xreg64

			timer [2]mmu_sc_portreg32

			timer_status [2]mmu_sc_xreg64

			timer_status_mask [2]mmu_sc_xreg64

			port_config mmu_sc_portreg32

			_ [0x15 - 0x0f]m.Reg32
		}

		// [0] 0 => werr, 1 => wrr
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

	// [16] enable write of [15:8] into port credit count
	// [15:8] credit count to write
	// [7:0] current credit count
	mmu_port_credit mmu_sc_portreg32

	_ [0x40 - 0x01]m.Reg32

	// per-port credit threshold for transition from store-and-forward to cut-through.
	// Default: 0x1
	asf_credit_threshold_hi mmu_sc_portreg32

	_ [0x80 - 0x41]m.Reg32

	mmu_1dbg_c mmu_sc_xreg32

	_ [1]m.Reg32

	mmu_dbg_c [3]mmu_sc_xreg32

	_ [0x90 - 0x85]m.Reg32

	mmu_1dbg_a mmu_sc_xreg32

	_ [0x38040000 - 0x38009100]byte

	// Port scheduler.
	tdm tdm_regs

	misc_config mmu_sc_reg32

	_ [1]m.Reg32

	// 9 pipe y tdm 1 in 4 error
	// 8 pipe x tdm 1 in 4 error
	// 7 start by start error
	// 6 pipe y detected deadlock
	// 5 pipe x detected deadlock
	// 4 cfap b bst trigger
	// 3 cfap a bst trigger
	// 2 memory parity error
	// 1 cfap b memory fail
	// 0 cfap a memory fail
	interrupt_enable mmu_sc_reg32
	interrupt_status mmu_sc_xreg32
	interrupt_clear  mmu_sc_xreg32

	_ [0x7 - 0x5]m.Reg32

	toq_multicast_config_0 mmu_sc_reg32

	toq_multicast_config_1 mmu_sc_xreg32

	toq_multicast_config_2 mmu_sc_xreg32

	start_by_start_error mmu_sc_xreg64
}

type mmu_sc_mems struct {
	_         [0x08040000 - 0x0]byte
	cell_link m.Mem
	pkt_link  m.Mem
	mcqe      m.Mem
	mcqn      m.Mem
	mcfp      m.Mem
	ucqdb_x   m.Mem
	mcqdb_x   [2]m.Mem // A/B
	ucqdb_y   m.Mem
	mcqdb_y   [2]m.Mem // a/b
	pdb       [2]m.Mem // X/Y
	_         [0x0c040000 - 0x08380000]byte

	replication struct {
		state      m.Mem
		group_info m.Mem
		head       m.Mem
		list       m.Mem
		_          [0x10040000 - 0x0c140000]byte
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
		_         [0x34000000 - 0x14200000]byte
	}

	queue_scheduler struct {
		l2_accumulated_compensation m.Mem
		l2_credit                   m.Mem
		l1_accumulated_compensation m.Mem
		l1_credit                   m.Mem
		l1_weight                   [n_tx_pipe]struct {
			unicast                  [n_mmu_data_port][mmu_n_tx_queues]m.Mem32
			unicast_cpu_management   [mmu_n_tx_queues]m.Mem32
			multicast                [n_mmu_data_port][mmu_n_tx_queues]m.Mem32
			unicast_loopback         [mmu_n_tx_queues]m.Mem32
			multicast_cpu_management [mmu_n_tx_queues]m.Mem32
			_                        [mmu_per_pipe_mem_bytes - (2*n_mmu_data_port+3)*mmu_n_tx_queues]m.MemElt
		}
		_                           [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt
		l0_accumulated_compensation m.Mem
		l0_credit                   m.Mem
		l0_weight                   [n_tx_pipe]struct {
			unicast [n_mmu_port][mmu_n_tx_queues]m.Mem32
			_       [mmu_per_pipe_mem_bytes - n_mmu_port*mmu_n_tx_queues]m.MemElt
		}
		_ [m.MemMax - n_pipe*mmu_per_pipe_mem_bytes]m.MemElt
	}

	_ [0x38000000 - 0x34200000]byte

	tdm_calendar [2]struct {
		entries [n_tx_pipe][mmu_per_pipe_mem_bytes]tdm_calendar_mem
		_       [m.MemMax - n_tx_pipe*mmu_per_pipe_mem_bytes]m.MemElt
	}

	_ [0x50000000 - 0x38080000]byte

	// [418:0]
	// [4 slices][15 banks]
	// within bank [4 xpe][index 0-1535 pad 1536-0x8000]
	// 4 slices makes 1 cell + parity/ecc
	cbp_data_slices [2]struct {
		entries [2][16][m.MemMax / mmu_per_pipe_mem_bytes][mmu_per_pipe_mem_bytes]mmu_cell_buffer_pool_mem
		_       [0x54000000 - 0x50800000]byte
	}
}
