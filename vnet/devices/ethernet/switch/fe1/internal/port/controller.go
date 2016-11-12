// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package port

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
)

type u32 m.U32

func (r *u32) get(q *dmaRequest, v *uint32) {
	(*m.U32)(r).Get(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}
func (r *u32) getDo(q *dmaRequest) (v uint32) { r.get(q, &v); q.Do(); return }

func (r *u32) set(q *dmaRequest, v uint32) {
	(*m.U32)(r).Set(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}

func (r *u32) offset() uint          { return (*m.U32)(r).Offset() }
func (r *u32) address() sbus.Address { return (*m.U32)(r).Address() }

type pu32 m.Pu32
type port_u32 [1 << m.Log2NPorts]pu32
type pu64 m.Pu64
type port_u64 [1 << m.Log2NPorts]pu64

func (r *pu32) get(q *dmaRequest, v *uint32) {
	(*m.Pu32)(r).Get(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}
func (r *pu32) getDo(q *dmaRequest) (v uint32) { r.get(q, &v); q.Do(); return }

func (r *pu32) set(q *dmaRequest, v uint32) {
	(*m.Pu32)(r).Set(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}

func (r *pu64) get(q *dmaRequest, v *uint64) {
	(*m.Pu64)(r).Get(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}
func (r *pu64) getDo(q *dmaRequest) (v uint64) { r.get(q, &v); q.Do(); return }

func (r *pu64) set(q *dmaRequest, v uint64) {
	(*m.Pu64)(r).Set(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}

func (r *pu32) offset() uint          { return (*m.Pu32)(r).Offset() }
func (r *pu32) address() sbus.Address { return (*m.Pu32)(r).Address() }
func (r *pu64) offset() uint          { return (*m.Pu64)(r).Offset() }
func (r *pu64) address() sbus.Address { return (*m.Pu64)(r).Address() }

type port_controller struct {
	counters [N_counters]port_u64

	_ [0x200 - N_counters]port_u64

	higig_config port_u32

	mib_stats_max_packet_size port_u32

	lag_failover_config port_u32

	eee_counter_mode port_u32

	_ u32

	link_status_to_cmic_control port_u32

	sw_flow_control port_u32

	flow_control_config u32

	mac_rsv_mask port_u32

	_ u32

	mode u32

	port_enable u32

	soft_reset u32

	power_save u32

	_ [0x210 - 0x20e]u32

	mac_control u32

	counter_mode u32

	_ u32

	phy_pll_status u32

	phy_control u32

	phy_lane_status [4]u32

	phy_uc_data_access_mode u32

	_ [0x224 - 0x21a]u32

	reset_mib_counters u32

	time_stamp_timer [2]u32

	link_status_down u32

	link_status_down_clear u32

	interrupt_status u32
	interrupt_enable u32

	sbus_control u32

	_ [0x600 - 0x22c]u32
}

type mac_common_0 struct {
	control port_u64

	mode port_u64

	spare [2]port_u64

	tx_control port_u64

	tx_src_address port_u64

	rx struct {
		control port_u64

		src_address port_u64

		max_bytes_per_packet port_u64

		ethernet_type_for_vlan port_u64

		lss_control port_u64

		lss_status       port_u64
		clear_lss_status port_u64
	}

	pause_control port_u64

	pfc struct {
		control     port_u64
		pfc_type    port_u64
		opcode      port_u64
		dst_address port_u64
	}

	llfc struct {
		control       port_u64
		tx_msg_fields port_u64
		rx_msg_fields port_u64
	}

	tx_timestamp_fifo_data   port_u64
	tx_timestamp_fifo_status port_u64

	fifo_status       port_u64
	fifo_status_clear port_u64

	lag_failover_status port_u64

	eee_control                 port_u64
	eee_timers                  port_u64
	eee_1_sec_link_status_timer port_u64

	higig_hdr [2]port_u64

	gmii_eee_control       port_u64
	tx_timestamp_adjust    port_u64
	tx_corrupt_crc_control port_u64

	e2e struct {
		control       port_u64
		cc_module_hdr [2]port_u64
		cc_data_hdr   [2]port_u64
		fc_module_hdr [2]port_u64
		fc_data_hdr   [2]port_u64
	}

	tx_fifo_cell_count port_u64

	tx_fifo_cell_request_count port_u64

	memory_control               port_u64
	ecc_control                  port_u64
	ecc_force_multiple_bit_error port_u64
	ecc_force_single_bit_error   port_u64
}

type mac_common_1 struct {
	rx_cdc_memory_ecc_status port_u64
	tx_cdc_memory_ecc_status port_u64
	ecc_status_clear         port_u64
	version                  port_u64
}

type hundred_gig_mac_uncommon_0 struct {
	rx_timestamp_memory_ecc_status port_u64
}

type hundred_gig_port_controller struct {
	port_controller
	mac_common_0
	hundred_gig_mac_uncommon_0
	mac_common_1
}

type forty_gig_port_controller struct {
	port_controller
	mac_common_0
	mac_common_1
}

func (p *PortBlock) get_controllers() (*port_controller, *mac_common_0, *mac_common_1, *hundred_gig_mac_uncommon_0) {
	if p.IsXlPort {
		x := (*forty_gig_port_controller)(m.BasePointer)
		return &x.port_controller, &x.mac_common_0, &x.mac_common_1, (*hundred_gig_mac_uncommon_0)(nil)
	} else {
		x := (*hundred_gig_port_controller)(m.BasePointer)
		return &x.port_controller, &x.mac_common_0, &x.mac_common_1, &x.hundred_gig_mac_uncommon_0
	}
}

type wc_ucmem_data_elt [4]uint32

func get_xclport_mems() *xclport_mems { return (*xclport_mems)(m.BasePointer) }

type xclport_mems struct {
	wc_ucmem_data m.Mem
}
