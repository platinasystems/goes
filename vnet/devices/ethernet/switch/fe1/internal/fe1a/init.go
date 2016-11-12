// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"fmt"
	"time"
)

func (t *fe1a) clear_mmu_packet_memory() {
	q := t.getDmaReq()
	zero := mmu_cell_buffer_pool_entry{}
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < mmu_n_cells_per_bank; k++ {
				for bank := 0; bank < mmu_n_banks; bank++ {
					for xpe := 0; xpe < 4; xpe++ {
						t.mmu_slice_mems.cbp_data_slices[i].entries[j][bank][xpe][k].set(q, &zero)
					}
				}
				q.Do()
			}
		}
	}
}

func (t *fe1a) clear_memories() {
	q := t.getDmaReq()

	// Initialize rx/tx pipe memories.
	{
		const (
			rx_pipe_done      uint32 = 1 << 21
			rx_pipe_valid     uint32 = 1 << 20
			rx_pipe_reset_all uint32 = 1 << 19
			tx_pipe_done      uint32 = 1 << 18
			tx_pipe_valid     uint32 = 1 << 17
			tx_pipe_reset_all uint32 = 1 << 16
		)

		// Initialize for largest memory size in rx pipe: ip4 fib tcam buckets
		count := uint32((n_shared_lookup_sram_bits_per_bucket / 70) * n_shared_lookup_sram_buckets_per_bank * n_shared_lookup_sram_banks)
		t.rx_pipe_controller.rx_buffer.hw_reset_control_1.set(q, rx_pipe_valid|rx_pipe_reset_all|count)

		count = uint32(len(t.tx_pipe_mems.l3_next_hop)) // next hops
		t.tx_pipe_controller.hw_reset_control_1.set(q, tx_pipe_valid|tx_pipe_reset_all|count)
		q.Do()

		// Wait for rx/tx pipe memory initialize to complete.
		// Timed to take ~1.5ms
		start := time.Now()
		iDone := 0
		eDone := 0
		var v [2 * n_pipe]uint32
		for (iDone & eDone) != 1<<n_pipe-1 {
			time.Sleep(200 * time.Microsecond)
			if time.Since(start) > 100*time.Millisecond {
				panic(fmt.Errorf("timeout waiting for rx_pipe/tx_pipe memory to initialize"))
			}

			for i := uint(0); i < n_pipe; i++ {
				if iDone&(1<<i) == 0 {
					t.rx_pipe_controller.rx_buffer.hw_reset_control_1.geta(q, sbus.Unique(i), &v[i])
				}
				if eDone&(1<<i) == 0 {
					t.tx_pipe_controller.hw_reset_control_1.geta(q, sbus.Unique(i), &v[n_pipe+i])
				}
			}

			q.Do()

			for i := uint(0); i < n_pipe; i++ {
				if v[i]&rx_pipe_done != 0 {
					iDone |= 1 << i
				}
				if v[n_pipe+i]&tx_pipe_done != 0 {
					eDone |= 1 << i
				}
			}
		}
	}

	// Initialize TDM calendar for rx scheduler.
	t.rx_pipe_controller.rx_buffer.tdm_calendar_init.set(q, 1)
	t.rx_pipe_controller.rx_buffer.tdm_calendar_init.set(q, 0)
	q.Do()

	// Initialize l2 management memories.
	{
		t.rx_pipe_controller.l2_management_hw_reset_control[0].seta(q, sbus.Unique0, 0)
		count := uint32(128) // L2 MOD FIFO
		t.rx_pipe_controller.l2_management_hw_reset_control[1].seta(q, sbus.Unique0, 1<<29|1<<30|count)
		q.Do()

		// Wait for done bit.
		start := time.Now()
		for {
			time.Sleep(200 * time.Microsecond)
			var v uint32
			t.rx_pipe_controller.l2_management_hw_reset_control[1].get(q, &v)
			q.Do()
			if v&(1<<31) != 0 {
				break
			}
			if time.Since(start) > 100*time.Millisecond {
				panic(fmt.Errorf("l2 mgmt memory initialization timeout"))
			}
		}
	}

	// Initialize mmu memories.
	{
		const (
			parity_enable uint32 = 1 << 2
			init_mem      uint32 = 1 << 1
		)

		v := uint32(parity_enable)
		t.mmu_global_controller.misc_config.set(q, v|init_mem)
		t.mmu_global_controller.misc_config.set(q, v)
		q.Do()
	}

	if false {
		t.clear_mmu_packet_memory()
	}
}

type obm_priority uint8

const (
	obm_priority_lossy_lo obm_priority = iota
	obm_priority_lossy_hi
	obm_priority_lossless_0
	obm_priority_lossless_1
	obm_n_priority
)

type obm_threshold_config struct {
	// Combined Lossy + Lossless > limit => discard.
	combined uint64
	// Lossless 0/1 limits.
	lossless [2]uint64
	// Combined lossy lo + hi > limit => discard
	lossy_0_plus_1               uint64
	lossy_0                      uint64
	lossy_0_plus_1_min_guarantee uint64
	// xoff when more cells than threshold; xon when less.
	combined_xoff uint64    // based on combined lossy + lossless
	lossless_xoff [2]uint64 // based on lossless 0/1
}

const n_obm_48byte_cells = 1024

var obm_settings = [...]obm_threshold_config{
	obm_threshold_config{},
	// 1 lane port
	obm_threshold_config{
		combined:       (n_obm_48byte_cells - 12) / 4,
		lossless:       [2]uint64{n_obm_48byte_cells - 1, n_obm_48byte_cells - 1},
		lossy_0_plus_1: 200,
		lossy_0:        100,
		lossy_0_plus_1_min_guarantee: 0,
		combined_xoff:                129,
		lossless_xoff:                [2]uint64{45, 45},
	},
	// 2 lane port
	obm_threshold_config{
		combined:       (n_obm_48byte_cells - 12) / 2,
		lossless:       [2]uint64{n_obm_48byte_cells - 1, n_obm_48byte_cells - 1},
		lossy_0_plus_1: 200,
		lossy_0:        100,
		lossy_0_plus_1_min_guarantee: 0,
		combined_xoff:                312,
		lossless_xoff:                [2]uint64{108, 108},
	},
	// 3 lane port (invalid)
	obm_threshold_config{},
	// 4 lane port
	obm_threshold_config{
		combined:       (n_obm_48byte_cells - 12) / 1,
		lossless:       [2]uint64{n_obm_48byte_cells - 1, n_obm_48byte_cells - 1},
		lossy_0_plus_1: 200,
		lossy_0:        100,
		lossy_0_plus_1_min_guarantee: 0,
		combined_xoff:                682,
		lossless_xoff:                [2]uint64{258, 258},
	},
}

func obm_for_port_block_index(i uint) (obm uint) {
	obm = i % 8
	if pipe := i / 8; pipe&1 != 0 {
		obm = 7 - obm
	}
	return
}

func (t *fe1a) rx_pipe_buffer_init() {
	type portBlock struct {
		speeds  [4]float64
		n_lanes [4]uint8
		pipe    uint
		n_ports uint
	}
	var pbs [n_port_block_100g]portBlock

	for _, p := range t.Ports {
		pc := p.GetPortCommon()
		if pc.IsManagement || !m.IsProvisioned(p) {
			continue
		}
		pb := pc.PortBlock
		bi := pb.GetPortBlockIndex()
		pi := m.GetSubPortIndex(p)
		pbs[bi].pipe = pb.GetPipe()
		pbs[bi].speeds[pi] = pc.SpeedBitsPerSec
		pbs[bi].n_lanes[pi] = uint8(p.GetLaneMask().NLanes())
		pbs[bi].n_ports++
	}

	q := t.getDmaReq()
	for port_block_index := range pbs {
		pb := &pbs[port_block_index]
		obm_index := obm_for_port_block_index(uint(port_block_index))
		port_mode := uint32(0)
		var ports []uint8
		switch pb.n_ports {
		case 1:
			port_mode = 0
			ports = []uint8{0}
		case 2:
			port_mode = 1
			ports = []uint8{0, 2}
		case 3:
			switch {
			case pb.speeds[2] == pb.speeds[3] && pb.speeds[0] >= 4*pb.speeds[2]:
				port_mode = 5 // 40G 10G 10G
				ports = []uint8{0, 2, 3}
			case pb.speeds[2] == pb.speeds[3] && pb.speeds[0] >= 2*pb.speeds[2]:
				port_mode = 3 // 20G 10G 10G
				ports = []uint8{0, 2, 3}
			case pb.speeds[0] == pb.speeds[1] && pb.speeds[2] >= 4*pb.speeds[0]:
				port_mode = 6 // 10G 10G 40G
				ports = []uint8{0, 1, 2}
			case pb.speeds[0] == pb.speeds[1] && pb.speeds[2] >= 2*pb.speeds[0]:
				port_mode = 4 // 10G 10G 20G
				ports = []uint8{0, 1, 2}
			default:
				panic("tri port")
			}
		case 4:
			port_mode = 1
			ports = []uint8{0, 1, 2, 3}
		default:
			panic("n ports")
		}
		// Set port mode and reset/unreset cell-assembly logic for all ports.
		v := port_mode << 4
		obm_regs := &t.rx_pipe_controller.over_subscription_buffer[obm_index]
		acc := sbus.Unique(pb.pipe)
		obm_regs.cell_assembly_control.seta(q, acc, v|0xf)
		obm_regs.cell_assembly_control.seta(q, acc, v|0x0)

		v = 0
		for i := range ports {
			// Bypass buffer if credits are available and obm is empty
			v |= 1 << (4*ports[i] + 1)
			// Enable over subscription mode.
			v |= 1 << (4*ports[i] + 0)
		}
		obm_regs.control.seta(q, acc, v)

		// Skip reset if port not over subscription mode.
		if false {
			continue
		}

		{
			const (
				discard_threshold       = 1023 << 0
				sop_threshold           = 1023 << 10
				sop_discard_mode_enable = 1 << 20
			)
			obm_regs.shared_config.seta(q, acc, uint32(discard_threshold|sop_threshold|sop_discard_mode_enable))
		}

		lossless := false
		if lossless {
			for pi := range ports {
				s := &obm_settings[pb.n_lanes[pi]]

				// Set port discard thresholds.
				v := s.combined << 50
				v |= s.lossless[1] << 40
				v |= s.lossless[0] << 30
				v |= s.lossy_0_plus_1 << 20
				v |= s.lossy_0 << 10
				v |= s.lossy_0_plus_1_min_guarantee << 0
				obm_regs.threshold[pi].seta(q, acc, v)

				// Set port flow control thresholds.
				v = (s.combined_xoff - 10) << 50
				v |= s.combined_xoff << 40
				v |= (s.lossless_xoff[1] - 10) << 30
				v |= s.lossless_xoff[1] << 20
				v |= (s.lossless_xoff[0] - 10) << 10
				v |= s.lossless_xoff[0] << 0
				obm_regs.flow_control_threshold[pi].seta(q, acc, v)

				// Enable port/lossless[01] flow control.  Cos bitmap for lossless01 packets: 0xff
				v = (1 << 34) | (1 << 33) | (1 << 32) | (0xff << 16) | (0xff << 0)
				obm_regs.flow_control_config[pi].seta(q, acc, v)

				// Default priority is lossless low (0).
				obm_regs.port_config[pi].seta(q, acc, uint32(obm_priority_lossless_0)<<0)

				// Map all 16 priorities to lossless low (0).
				w := uint32(0)
				for i := uint(0); i < 16; i++ {
					w |= uint32(obm_priority_lossless_0) << (2 * i)
				}
				t.rx_pipe_mems.over_subscription_buffer[obm_index].priority_map[pi][0].Set(&q.DmaRequest, BlockRxPipe, acc, w)
			}

			// Select lossless low (0) for max usage counter.
			obm_regs.max_usage_select.seta(q, acc, uint32(obm_priority_lossless_0)<<0)
		} else {
			for pi := range ports {
				s := &obm_settings[pb.n_lanes[pi]]

				// Set port discard thresholds.
				v := s.combined << 50
				v |= s.lossless[1] << 40
				v |= s.lossless[0] << 30
				v |= s.lossy_0_plus_1 << 20
				v |= s.lossy_0 << 10
				v |= s.lossy_0_plus_1_min_guarantee << 0
				obm_regs.threshold[pi].seta(q, acc, v)
			}
		}
	}

	// Put cpu/loopback cell assembly in reset and release to send initial credit to tx_pipe.
	t.rx_pipe_controller.rx_buffer.cell_assembly_cpu.control.set(q, 1)
	t.rx_pipe_controller.rx_buffer.cell_assembly_loopback.control.set(q, 1)
	t.rx_pipe_controller.rx_buffer.cell_assembly_cpu.control.set(q, 0)
	t.rx_pipe_controller.rx_buffer.cell_assembly_loopback.control.set(q, 0)

	q.Do()
}

func (t *fe1a) port_bitmap_garbage_dump_init() {
	q := t.getDmaReq()

	// Initialize CPU port bitmaps.
	{
		c := make_port_bitmap_entry(t.cpu_ports)
		t.rx_pipe_mems.cpu_port_bitmap[0].seta(q, &c, BlockRxPipe, sbus.Duplicate)
		t.rx_pipe_mems.cpu_port_bitmap_1[0].seta(q, &c, BlockRxPipe, sbus.Duplicate)
		q.Do()
	}

	// Initialize loopback port bitmaps.
	{
		for p := uint(0); p < n_pipe; p++ {
			d := &port_bitmap_entry{}
			d.add(phys_port_loopback_for_pipe(p).toPipe())
			t.rx_pipe_mems.multipass_loopback_bitmap[0].seta(q, d, BlockRxPipe, sbus.Unique(p))
		}
		q.Do()
	}

	// Set tx_pipe port type for loopback ports.
	{
		const (
			loopback = 2
			hi_gig   = 1
		)
		for pipe := uint(0); pipe < n_pipe; pipe++ {
			pipePort := phys_port_loopback_for_pipe(pipe).toPipe()
			t.tx_pipe_mems.rx_port[pipePort].Set(&q.DmaRequest, BlockTxPipe, sbus.Duplicate, loopback)
		}
		t.tx_pipe_mems.rx_port[136].Set(&q.DmaRequest, BlockTxPipe, sbus.Duplicate, hi_gig)
		q.Do()
	}

	// MMU refresh clock enable.
	{
		v := t.mmu_global_controller.misc_config.getDo(q)
		v |= 1 << 0
		t.mmu_global_controller.misc_config.set(q, v)
		q.Do()
	}

	// Hash init.
	{
		// 64 bit Hash Vector: {hash_zero[15:0] or hash_lsb[15:0], crc16[15:0], crc32[31:0]}

		// l2_entry table: 1k entries x 2 banks
		// bank0 crc32[15:0] bank1 crc32[31:16]
		t.rx_pipe_controller.l2_table_hash_control.set(q, 0<<(6*0)|16<<(6*1))

		// l3_entry table: 512 entries x 4 banks
		// banks 6-9
		t.rx_pipe_controller.l3_table_hash_control.set(q, (9*0)<<(6*0)|(9*1)<<(6*1)|(9*2)<<(6*2)|(9*3)<<(6*3))

		// shared iss: 8k entries x 4 banks
		// banks 2-5
		v := uint32(0xf << 24) // enable all 4 banks
		v |= 0<<(6*0) | 6<<(6*1) | 13<<(6*2) | 19<<(6*3)
		t.rx_pipe_controller.shared_table_hash_control.set(q, v)
		q.Do()
	}

	t.tx_pipe_mmu_credits_set()

	// Set cut-through/store-and-forward tx_pipe transmit count.
	{
		for phys := physical_port_number(0); phys < n_phys_ports; phys++ {
			pipePort := phys.toPipe()
			if t.all_ports.isSet(pipePort) || t.loopback_ports.isSet(pipePort) {
				_, mmu_pipe := phys.toMmu(t)
				pi := pipePort % 34
				for class := 0; class < n_asf_class; class++ {
					// units of 16 bytes => 288 bytes
					v := uint32(18)
					t.tx_pipe_mems.tx_start_count[pi][class].Set(&q.DmaRequest, BlockTxPipe, sbus.Unique(uint(mmu_pipe)), v)
				}
				q.Do()
			}
		}
	}

	// High speed port multicast t2oq settings.
	{
		var vs [2]uint32
		for i := range t.port_by_phys_port {
			if p := t.port_by_phys_port[i]; p != nil && p.SpeedBitsPerSec >= 40e9 {
				m, pipe := p.physical_port_number.toMmu(t)
				slice := pipe / 2
				shift := uint(m & 0xf)
				if pipe&1 != 0 {
					shift += 16
				}
				vs[slice] |= 1 << shift
			}
		}
		for slice := range vs {
			t.mmu_slice_controller.toq_multicast_config_1[slice].set(q, mmuBaseTypeSlice, sbus.Single, vs[slice])
		}
		// Set MCQE_FULL_THRESHOLD to 0.
		t.mmu_slice_controller.toq_multicast_config_0.set(q, mmuBaseTypeChip, sbus.Duplicate, 0)
		q.Do()
	}

	// Set mysterious undocumented edb memory.
	{
		for phys := physical_port_number(0); phys < n_phys_ports; phys++ {
			pipePort := phys.toPipe()
			if t.all_ports.isSet(pipePort) {
				speed := phys.speedBitsPerSec(t)
				b := uint32(11875 * speed / 100e9)
				t.tx_pipe_mems.data_buffer_1dbg_b[phys].Set(&q.DmaRequest, BlockTxPipe, sbus.AddressSplitDist, b)
			}
		}

		{
			a := uint32(t.CoreFrequencyInHz * 1e-6)
			b := uint32(0)
			switch t.CoreFrequencyInHz {
			case 850e6:
				b = 7650
			case 765e6:
				b = 6885
			case 672e6:
				b = 6048
			case 645e6:
				b = 5805
			case 545e6:
				b = 4905
			default:
				panic("unsupported frequency")
			}
			c := uint32(8)
			v := uint32(a<<0 | b<<10 | c<<24)
			for tx_pipe := uint(0); tx_pipe < n_tx_pipe; tx_pipe++ {
				t.tx_pipe_controller.data_buffer_1dbg_a.seta(q, sbus.Unique(tx_pipe), v)
			}
		}
		q.Do()
	}

	{
		for phys := physical_port_number(0); phys < n_phys_ports; phys++ {
			pipePort := phys.toPipe()
			if !(t.all_ports.isSet(pipePort) || t.loopback_ports.isSet(pipePort)) {
				continue
			}
			mmuPort := phys.toGlobalMmu(t)
			t.tx_pipe_mems.per_port_buffer_soft_reset[pipePort].Set(&q.DmaRequest, BlockTxPipe, sbus.AddressSplit, 1)
			if mmuPort&0x3f < 32 {
				asf_class := uint32(asf_100g)
				t.tx_pipe_mems.ip_cut_thru_class[pipePort].Set(&q.DmaRequest, BlockTxPipe, sbus.Duplicate, asf_class)
				t.mmu_slice_controller.asf_iport_config[mmuPort].set(q, mmuBaseTypeRxPort, sbus.Duplicate, asf_class)
				t.mmu_slice_controller.asf_credit_threshold_hi[mmuPort].set(q, mmuBaseTypeTxPort, sbus.Single, 0x1b)
			}
			t.mmu_slice_controller.mmu_port_credit[mmuPort].set(q, mmuBaseTypeTxPort, sbus.Single, 1<<16)
			t.mmu_slice_controller.mmu_port_credit[mmuPort].set(q, mmuBaseTypeTxPort, sbus.Single, 0)
			t.tx_pipe_mems.per_port_buffer_soft_reset[pipePort].Set(&q.DmaRequest, BlockTxPipe, sbus.AddressSplit, 0)
			t.set_cell_assembly_cut_through_threshold(q, phys, 4)
		}
		q.Do()
	}

	// Enable egress to request cells from mmu for all ports.
	{
		for phys := physical_port_number(0); phys < n_phys_ports; phys++ {
			pipePort := phys.toPipe()
			if t.all_ports.isSet(pipePort) || t.loopback_ports.isSet(pipePort) {
				const enable = 1
				t.tx_pipe_mems.port_enable[pipePort].Set(&q.DmaRequest, BlockTxPipe, sbus.AddressSplit, enable)
			}
		}
		q.Do()
	}

	// EPC link bitmap.  Cpu port always has "link" up.
	{
		c := make_port_bitmap_entry(t.cpu_ports)
		// Enable SOBMH blocking bit.
		c.add(pipe_port_number(136))
		c.or(&t.loopback_ports)
		c.or(&t.all_ports)
		t.rx_pipe_mems.epc_link_port_bitmap[0].seta(q, &c, BlockRxPipe, sbus.Duplicate)
		q.Do()
	}

	// Rx pipe config.
	{
		v := t.rx_pipe_controller.rx_config.getDo(q, sbus.Duplicate)

		const (
			apply_mask_on_l2_packets          = 1 << 12
			apply_mask_on_l3_packets          = 1 << 13
			arp_included_in_rxf_keys          = 1 << 39
			rarp_included_in_rxf_keys         = 1 << 40
			arp_validation_enable             = 1 << 29
			ignore_hi_gig_header_lag_failover = 1 << 23
		)

		v |= apply_mask_on_l3_packets | apply_mask_on_l2_packets

		v |= arp_included_in_rxf_keys | rarp_included_in_rxf_keys | arp_validation_enable

		if false {
			v |= ignore_hi_gig_header_lag_failover
		}

		t.rx_pipe_controller.rx_config.set(q, v)
		q.Do()
	}

	{
		v := t.tx_pipe_controller.config_1.getDo(q, sbus.Duplicate)
		// Enable ring mode
		v |= uint32(1 << 0)
		t.tx_pipe_controller.config_1.set(q, v)
		q.Do()
	}

	{
		v := uint32(0)
		v |= 0 << 3  // disable force vlan translation misses to be untagged.
		v |= 1 << 11 // enable pri/cfi remarking.
		for p := pipe_port_number(0); p < n_pipe_ports; p++ {
			t.tx_pipe_controller.vlan_control_1[p].seta(q, sbus.AddressSplit, v)
		}
		q.Do()
	}

	{
		c := make_port_bitmap_entry(t.all_ports)
		t.rx_pipe_mems.vlan_membership_check_enable_port_bitmap[0].seta(q, &c, BlockRxPipe, sbus.Duplicate)
		q.Do()
	}

	t.CpuMain.MdioInit(t.CoreFrequencyInHz, t)

	// Ports start at data ram offset 0xa0 on fe1a.
	t.CpuMain.LedInit(0xa0)

	// Re-program eprg kill timeout
	{
		v := uint32(512) | (1 << 10)
		for slice := 0; slice < 2; slice++ {
			t.mmu_slice_controller.toq_multicast_config_2[slice].set(q, mmuBaseTypeSlice, sbus.Single, v)
		}
	}
	q.Do()
}

func (t *fe1a) tx_pipe_mmu_credits_set() {
	q := t.getDmaReq()

	tab_850mhz := [n_port_speed]uint32{
		port_speed_lt_20g:  11,
		port_speed_lt_25g:  16,
		port_speed_lt_40g:  15,
		port_speed_lt_50g:  19,
		port_speed_lt_100g: 23,
		port_speed_ge_100g: 36,
	}
	tab_other := [n_port_speed]uint32{
		port_speed_lt_20g:  13,
		port_speed_lt_25g:  18,
		port_speed_lt_40g:  16,
		port_speed_lt_50g:  25,
		port_speed_lt_100g: 27,
		port_speed_ge_100g: 44,
	}
	tab := tab_850mhz[:]
	if t.CoreFrequencyInHz != 850e6 {
		tab = tab_other[:]
	}

	for phys := physical_port_number(0); phys < n_phys_ports; phys++ {
		var (
			v  uint32
			ok bool
		)

		switch {
		case phys.is_loopback_port():
			v = tab[port_speed_ge_100g]
			ok = true
		case phys == phys_port_cpu:
			v = tab[port_speed_lt_20g]
			ok = true
		default:
			if port := t.port_by_phys_port[phys]; port != nil && m.IsProvisioned(port) {
				v = tab[port_speed_code(port.SpeedBitsPerSec)]
				ok = true
			}
		}
		if ok {
			pipePort := physical_port_number(phys).toPipe()
			t.tx_pipe_controller.mmu_max_cell_credit[pipePort].seta(q, sbus.AddressSplit, v)
		}
	}
	q.Do()
}

func (t *fe1a) misc_init() {
	t.port_mapping_init()
	t.tdm_scheduler_init()
	t.rx_pipe_buffer_init()
	t.port_bitmap_garbage_dump_init()
	t.mmu_init()
	t.tmon_init()
}
