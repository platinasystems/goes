// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"math/rand"
)

func (t *fe1a) mmu_init() {
	q := t.getDmaReq()

	max_packet_cells := bytesToCells(mmu_max_packet_bytes)

	{
		enter := uint32(mmu_n_cells_per_pipe - mmu_reserved_cfap_cells)
		exit := enter - 2*uint32(max_packet_cells)
		for p := 0; p < n_mmu_pipe; p++ {
			t.mmu_slice_controller.cfap.enter_full_threshold[p].set(q, mmuBaseTypeMmuPipe, sbus.Single, enter)
			t.mmu_slice_controller.cfap.exit_full_threshold[p].set(q, mmuBaseTypeMmuPipe, sbus.Single, exit)
		}
		q.Do()
	}

	// For all ports, map: priority 0 => priority group 0
	//                     others => priority group 1
	// creating a Hi and Lo priority scheme.
	// and set input port rx enable, but no xon or pause frame support.
	{
		var v [2]uint32

		for pri := uint32(0); pri < 16; pri++ {
			pg := uint32(0)
			if pri > 0 {
				pg = 1
			}
			i0, i1 := pri/8, 3*(pri%8)
			v[i0] |= pg << i1
		}
		for p := 0; p < n_mmu_port; p++ {
			for i := 0; i < 2; i++ {
				t.mmu_pipe_controller.rx_admission_control.port_priority_group[i][p].set(q, mmuBaseTypeRxPort, sbus.Duplicate, v[i])
			}
		}
		q.Do()
	}

	global_headroom_cells := max_packet_cells

	{
		// Allow headroom for 1 max packet cells.
		for p := 0; p < n_pipe; p++ {
			t.mmu_pipe_controller.rx_admission_control.global_headroom_limit[p].set(q, mmuBaseTypeRxPipe, sbus.Duplicate, uint32(global_headroom_cells))
		}

		// Subtract one since first cell of packet is never stored in global headroom.
		t.mmu_pipe_controller.rx_admission_control.max_packet_cells.set(q, mmuBaseTypeChip, sbus.Duplicate, uint32(max_packet_cells-1))
		q.Do()
	}

	{
		// Configure the Ingress Process Group settings
		// Currently not dynamic, and setup as Hi and Lo priority pools
		// for the traffic priority (Hi: priority/process group 0, otherwise Lo:)
		total_cells := float64(mmu_n_cells_per_pipe - mmu_reserved_cfap_cells - global_headroom_cells)
		cells_left := total_cells

		const n_rx_port_per_mmu_pipe = 16

		// Hi priority traffic gets dedicated service pool.
		per_port_hi_cells := .02 * total_cells
		cells_left -= per_port_hi_cells
		per_port_hi_cells /= n_rx_port_per_mmu_pipe

		// Allocate 3% of total cells as guaranteed split evenly among data ports.
		per_port_lo_cells := .03 * total_cells
		cells_left -= per_port_lo_cells
		per_port_lo_cells /= n_rx_port_per_mmu_pipe

		pgs := [...]mmu_priority_group_config_entry{
			0: mmu_priority_group_config_entry{
				shared_limit:            mmu_cell_count(total_cells),
				min_reserved:            mmu_cell_count(per_port_lo_cells),
				global_headroom_enable:  true,
				shared_limit_is_dynamic: false,
			},
			1: mmu_priority_group_config_entry{
				shared_limit:            mmu_cell_count(total_cells),
				min_reserved:            mmu_cell_count(per_port_hi_cells),
				global_headroom_enable:  true,
				shared_limit_is_dynamic: false,
			},
		}

		// assign the mmu priority group pool entries: all ports (3%) and hi priority (2%)
		for pgi := range pgs {
			for port := 0; port < n_mmu_port; port++ {
				for pipe := uint(0); pipe < n_rx_pipe; pipe++ {
					t.mmu_pipe_mems.rx_admission_control.port_priority_group_config[pipe].entries[port][pgi].set(q, &pgs[pgi])
				}
			}
		}
		q.Do()
	}

	rx_pool_shared_limit := uint32(mmu_n_cells_per_pipe - mmu_reserved_cfap_cells - global_headroom_cells)

	rx_service_pools := [...]mmu_rx_service_pool_config_entry{
		0: mmu_rx_service_pool_config_entry{
			limit:        mmu_cell_count(rx_pool_shared_limit),
			min_reserved: 0,
			resume_limit: mmu_cell_count(rx_pool_shared_limit - 16),
		},
	}

	// assign the mmu shared pool entry
	{
		for pi := range rx_service_pools {
			p := &rx_service_pools[pi]
			for layer := 0; layer < n_mmu_layer; layer++ {
				// Packets dropped when limit is reached.  Drop state is released at limit - reset offset.
				t.mmu_pipe_controller.rx_admission_control.service_pool_shared_cell_limit[pi][layer].set(q, mmuBaseTypeLayer, sbus.Duplicate, uint32(p.limit))
				t.mmu_pipe_controller.rx_admission_control.service_pool_cell_reset_limit_offset[pi][layer].set(q, mmuBaseTypeLayer, sbus.Duplicate, uint32(16))
			}

			for port := 0; port < n_mmu_port; port++ {
				for pipe := uint(0); pipe < n_rx_pipe; pipe++ {
					t.mmu_pipe_mems.rx_admission_control.port_service_pool_config[pipe].entries[port][pi].set(q, p)
				}
			}
		}
		q.Do()
	}

	{
		for phys := physical_port_number(0); phys < n_phys_ports; phys++ {
			pipePort := phys.toPipe()
			if t.all_ports.isSet(pipePort) || t.loopback_ports.isSet(pipePort) {
				mmuPort := phys.toGlobalMmu(t)
				// Input port rx enable; no pause frames or priority xon enabled.
				const v = 1 << 17
				t.mmu_pipe_controller.rx_admission_control.port_rx_and_pause_enable[mmuPort].set(q, mmuBaseTypeRxPort, sbus.Duplicate, v)
			}
		}
		q.Do()
	}

	// Egress service pools apply to both unicast & multicast.
	t.mmu_pipe_controller.multicast_admission_control.db.service_pool_shared_limit[0].set(q, mmuBaseTypeChip, sbus.Duplicate, rx_pool_shared_limit)
	t.mmu_pipe_controller.multicast_admission_control.mcqe.pool_shared_limit[0].set(q, mmuBaseTypeChip, sbus.Duplicate, mmu_n_mcqe/4-1)
	t.mmu_pipe_controller.db.service_pool_config[0].set(q, mmuBaseTypeChip, sbus.Duplicate, rx_pool_shared_limit)
	t.mmu_pipe_controller.qe.service_pool_config[0].set(q, mmuBaseTypeChip, sbus.Duplicate, uint32(mmu_n_rqe)/8-1)
	for i := range t.mmu_pipe_controller.db.config_1 {
		t.mmu_pipe_controller.db.config_1[i].set(q, mmuBaseTypeChip, sbus.Duplicate, rx_pool_shared_limit)
		t.mmu_pipe_controller.qe.config_1[i].set(q, mmuBaseTypeChip, sbus.Duplicate, uint32(mmu_n_rqe)/8-1)
	}
	q.Do()

	{
		const (
			enable_queue_and_group_ticket = 1 << 4
			mop_policy_1b                 = 1 << 1
			mop_policy                    = 1 << 0
		)
		t.mmu_pipe_controller.tx_admission_control.config.set(q, mmuBaseTypeChip, sbus.Duplicate, uint32(enable_queue_and_group_ticket|mop_policy_1b))
		t.mmu_pipe_controller.multicast_admission_control.db.config.set(q, mmuBaseTypeChip, sbus.Duplicate, mop_policy)
		t.mmu_pipe_controller.db.config.set(q, mmuBaseTypeChip, sbus.Duplicate, (1<<2)|mop_policy_1b)
		q.Do()
	}

	// Unicast port service pools.
	for pi := range rx_service_pools {
		p := &rx_service_pools[pi]
		for port := 0; port < n_mmu_port; port++ {
			for pipe := uint(0); pipe < n_tx_pipe; pipe++ {
				y := mmu_8cell_count(p.limit / 8)
				x := mmu_tx_service_pool_config_entry{
					shared_limit: p.limit,
					yellow_limit: y,
					red_limit:    y,
				}
				t.mmu_pipe_mems.tx_admission_control.service_pool_config[pipe].entries[port][pi].set(q, &x)

				l := mmu_8cell_count((p.limit - p.resume_limit) / 8)

				r := mmu_tx_color_config_entry{
					green:  l,
					yellow: l,
					red:    l,
				}
				t.mmu_pipe_mems.tx_admission_control.resume_config[pipe].entries[port][pi].set(q, &r)
			}
		}
	}
	q.Do()

	// Multicast port service pools.
	{
		e := mmu_multicast_db_service_pool_config_entry{
			shared_limit_enable: true,
			shared_limit:        mmu_cell_count(rx_pool_shared_limit),
		}
		f := mmu_multicast_mcqe_service_pool_config_entry{
			shared_limit_enable: true,
			shared_limit:        mmu_4mcqe_count(mmu_n_mcqe/4 - 1),
		}

		for port := 0; port < n_mmu_port; port++ {
			for pipe := uint(0); pipe < n_tx_pipe; pipe++ {
				t.mmu_pipe_mems.multicast_admission_control.db_port_service_pool_config[pipe].ports[port][0].set(q, &e)
				t.mmu_pipe_mems.multicast_admission_control.mcqe_port_service_pool_config[pipe].ports[port][0].set(q, &f)
			}
		}
		q.Do()
	}

	{
		queue_configs := [...]mmu_tx_queue_config_entry{
			0: mmu_tx_queue_config_entry{
				min_reserved: 0,
				shared_limit: rx_service_pools[0].limit,
			},
		}
		for pipe := uint(0); pipe < n_tx_pipe; pipe++ {
			for qi := range queue_configs {
				for port := 0; port < n_mmu_data_port; port++ {
					t.mmu_pipe_mems.tx_admission_control.queue_config[pipe].data_port_entries[port][qi].set(q, &queue_configs[qi])
				}
				for port := 0; port < 2; port++ {
					t.mmu_pipe_mems.tx_admission_control.queue_config[pipe].cpu_loopback_port_entries[port][qi].set(q, &queue_configs[qi])
				}
			}
		}
		q.Do()
	}

	{
		const weight = 1
		for pipe := uint(0); pipe < n_tx_pipe; pipe++ {
			for tq := 0; tq < mmu_n_tx_queues; tq++ {
				for port := 0; port < n_mmu_data_port; port++ {
					t.mmu_slice_mems.queue_scheduler.l1_weight[pipe].unicast[port][tq].Seta(&q.DmaRequest, BlockMmuSlice, sbus.Single, mmuBaseTypeTxPipe, weight)
					t.mmu_slice_mems.queue_scheduler.l1_weight[pipe].multicast[port][tq].Seta(&q.DmaRequest, BlockMmuSlice, sbus.Single, mmuBaseTypeTxPipe, weight)
				}
				t.mmu_slice_mems.queue_scheduler.l1_weight[pipe].unicast_cpu_management[tq].Seta(&q.DmaRequest, BlockMmuSlice, sbus.Single, mmuBaseTypeTxPipe, weight)
				t.mmu_slice_mems.queue_scheduler.l1_weight[pipe].multicast_cpu_management[tq].Seta(&q.DmaRequest, BlockMmuSlice, sbus.Single, mmuBaseTypeTxPipe, weight)
				t.mmu_slice_mems.queue_scheduler.l1_weight[pipe].unicast_loopback[tq].Seta(&q.DmaRequest, BlockMmuSlice, sbus.Single, mmuBaseTypeTxPipe, weight)
				for port := 0; port < n_mmu_port; port++ {
					t.mmu_slice_mems.queue_scheduler.l0_weight[pipe].unicast[port][tq].Seta(&q.DmaRequest, BlockMmuSlice, sbus.Single, mmuBaseTypeTxPipe, weight)
				}
			}
			q.Do()
		}
	}

	{
		// Set undocumented a/c fields
		for pipe := uint(0); pipe < n_tx_pipe; pipe++ {
			var v uint32
			t.mmu_slice_controller.mmu_1dbg_c[pipe].get(q, mmuBaseTypeTxPipe, sbus.Single, &v)
			q.Do()

			v |= 1 << 0
			t.mmu_slice_controller.mmu_1dbg_c[pipe].set(q, mmuBaseTypeTxPipe, sbus.Single, v)
			t.mmu_slice_controller.mmu_1dbg_a[pipe].set(q, mmuBaseTypeTxPipe, sbus.Single, ^uint32(0))
			q.Do()
		}
	}

	{
		for _, porter := range t.Ports {
			p := porter.(*Port)
			v := uint32(15)
			switch {
			case p.SpeedBitsPerSec >= 100e9:
				v = 140
			case p.SpeedBitsPerSec >= 40e9:
				v = 60
			case p.SpeedBitsPerSec >= 25e9:
				v = 40
			case p.SpeedBitsPerSec >= 20e9:
				v = 30
			}
			v += uint32(rand.Intn(20))
			i := p.physical_port_number.toGlobalMmu(t)
			if (i & 0x3f) != 32 { // cpu/management ports cause crash.
				t.mmu_slice_controller.mmu_dbg_c[2][i].set(q, mmuBaseTypeTxPort, sbus.Single, v)
			}
		}
		q.Do()
	}

	{
		var v [n_tx_pipe]uint64
		for phys := physical_port_number(0); phys < n_phys_ports; phys++ {
			pipePort := phys.toPipe()
			if t.all_ports.isSet(pipePort) || t.loopback_ports.isSet(pipePort) {
				mmu_port, mmu_pipe := phys.toMmu(t)
				v[mmu_pipe] |= uint64(1) << mmu_port
			}
		}
		for p := uint(0); p < n_tx_pipe; p++ {
			pipe_mask := mmu_pipe_mask_for_tx_pipe(p)
			for mmu_pipe := uint(0); mmu_pipe < n_mmu_pipe; mmu_pipe++ {
				if pipe_mask&(1<<mmu_pipe) != 0 {
					t.mmu_pipe_controller.tx_admission_control.tx_port_enable[p].set(q, mmuBaseTypeTxPipe, sbus.Unique(mmu_pipe), v[p])
					t.mmu_pipe_controller.multicast_admission_control.db.port_tx_enable[p].set(q, mmuBaseTypeTxPipe, sbus.Unique(mmu_pipe), v[p])
					t.mmu_pipe_controller.multicast_admission_control.mcqe.port_tx_enable[p].set(q, mmuBaseTypeTxPipe, sbus.Unique(mmu_pipe), v[p])
				}
			}
		}
		q.Do()
	}

	if false {
		t.mmu_pipe_controller.tx_admission_control.bypass.set(q, mmuBaseTypeChip, sbus.Duplicate, 1)
		q.Do()
	}
}

type mmu_rx_counter_entry vnet.CombinedCounter
type mmu_tx_counter_entry vnet.CombinedCounter
type mmu_wred_counter_entry uint64

type mmu_rx_counter_mem m.MemElt
type mmu_tx_counter_mem m.MemElt
type mmu_wred_counter_mem m.MemElt

func (e *mmu_rx_counter_entry) MemBits() int { return 71 }
func (e *mmu_rx_counter_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint64(&e.Packets, b, i+31, i, isSet)
	i = m.MemGetSetUint64(&e.Bytes, b, i+37, i, isSet)
}

func (e *mmu_tx_counter_entry) MemBits() int { return 81 }
func (e *mmu_tx_counter_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint64(&e.Packets, b, i+35, i, isSet)
	i = m.MemGetSetUint64(&e.Bytes, b, i+43, i, isSet)
}

func (e *mmu_wred_counter_entry) MemBits() int { return 37 }
func (e *mmu_wred_counter_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint64((*uint64)(e), b, i+35, i, isSet)
}

func (e *mmu_rx_counter_mem) get(q *DmaRequest, v *mmu_rx_counter_entry) {
	(*m.MemElt)(e).MemDmaGeta(&q.DmaRequest, v, BlockMmuPipe, sbus.Single, mmuBaseTypeMmuPipe)
}
func (e *mmu_rx_counter_mem) set(q *DmaRequest, v *mmu_rx_counter_entry) {
	(*m.MemElt)(e).MemDmaSeta(&q.DmaRequest, v, BlockMmuPipe, sbus.Single, mmuBaseTypeMmuPipe)
}

func (e *mmu_tx_counter_mem) get(q *DmaRequest, mmu_pipe uint, v *mmu_tx_counter_entry) {
	(*m.MemElt)(e).MemDmaGeta(&q.DmaRequest, v, BlockMmuPipe, sbus.Unique(mmu_pipe), mmuBaseTypeTxPipe)
}
func (e *mmu_tx_counter_mem) set(q *DmaRequest, mmu_pipe uint, v *mmu_tx_counter_entry) {
	(*m.MemElt)(e).MemDmaSeta(&q.DmaRequest, v, BlockMmuPipe, sbus.Unique(mmu_pipe), mmuBaseTypeTxPipe)
}

func (e *mmu_wred_counter_mem) get(q *DmaRequest, mmu_pipe uint, v *mmu_wred_counter_entry) {
	(*m.MemElt)(e).MemDmaGeta(&q.DmaRequest, v, BlockMmuPipe, sbus.Unique(mmu_pipe), mmuBaseTypeTxPipe)
}
func (e *mmu_wred_counter_mem) set(q *DmaRequest, mmu_pipe uint, v *mmu_wred_counter_entry) {
	(*m.MemElt)(e).MemDmaSeta(&q.DmaRequest, v, BlockMmuPipe, sbus.Unique(mmu_pipe), mmuBaseTypeTxPipe)
}

// ASF: alternate store and forward mode (cut through)

type asf_class uint8

const (
	asf_store_and_forward asf_class = iota
	asf_10g
	asf_11g // hi-gig
	asf_20g
	asf_21g // hi-gig
	asf_25g
	asf_27g // hi-gig
	asf_40g
	asf_42g // hi-gig
	asf_50g
	asf_53g // hi-gig
	asf_100g
	asf_106g    // hi-gig
	n_asf_class = 13
)

func (t *fe1a) set_cell_assembly_cut_through_threshold(q *DmaRequest, p physical_port_number, thresh uint) {
	idb, rx_tx_pipe := p.to_rx_pipe_mmu()
	pipe := uint(rx_tx_pipe)
	i0, i1 := uint(idb/4), uint(idb%4)
	port_block_index := 8*pipe + i0
	obm := obm_for_port_block_index(port_block_index)
	var v uint32
	t.rx_pipe_controller.over_subscription_buffer[obm].cell_assembly_cut_through_control.geta(q, sbus.Unique(pipe), &v)
	q.Do()
	v = (v &^ (3 << (20 + 2*i1))) | (uint32(i1) << (20 + 2*i1))
	v = (v &^ (0x1f << (0 + 5*i1))) | ((uint32(thresh) & 0x1f) << (0 + 5*i1))
	t.rx_pipe_controller.over_subscription_buffer[obm].cell_assembly_cut_through_control.seta(q, sbus.Unique(pipe), v)
	q.Do()
}
