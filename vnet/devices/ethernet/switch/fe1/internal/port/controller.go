// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package port

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
)

type reg32 m.U32

func (r *reg32) get(q *dmaRequest, v *uint32) {
	(*m.U32)(r).Get(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}
func (r *reg32) getDo(q *dmaRequest) (v uint32) { r.get(q, &v); q.Do(); return }

func (r *reg32) set(q *dmaRequest, v uint32) {
	(*m.U32)(r).Set(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}

func (r *reg32) offset() uint          { return (*m.U32)(r).Offset() }
func (r *reg32) address() sbus.Address { return (*m.U32)(r).Address() }

type preg32 m.Pu32
type portreg32 [1 << m.Log2NRegPorts]preg32
type preg64 m.Pu64
type portreg64 [1 << m.Log2NRegPorts]preg64

func (r *preg32) get(q *dmaRequest, v *uint32) {
	(*m.Pu32)(r).Get(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}
func (r *preg32) getDo(q *dmaRequest) (v uint32) { r.get(q, &v); q.Do(); return }

func (r *preg32) set(q *dmaRequest, v uint32) {
	(*m.Pu32)(r).Set(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}

func (r *preg64) get(q *dmaRequest, v *uint64) {
	(*m.Pu64)(r).Get(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}
func (r *preg64) getDo(q *dmaRequest) (v uint64) { r.get(q, &v); q.Do(); return }

func (r *preg64) set(q *dmaRequest, v uint64) {
	(*m.Pu64)(r).Set(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}

func (r *preg32) offset() uint          { return (*m.Pu32)(r).Offset() }
func (r *preg32) address() sbus.Address { return (*m.Pu32)(r).Address() }
func (r *preg64) offset() uint          { return (*m.Pu64)(r).Offset() }
func (r *preg64) address() sbus.Address { return (*m.Pu64)(r).Address() }

type xclport_regs struct {
	counters [N_counters]portreg64

	_ [0x200 - N_counters]portreg64

	higig_config portreg32

	mib_stats_max_packet_size portreg32

	lag_failover_config portreg32

	eee_counter_mode portreg32

	_ reg32

	link_status_to_cmic_control portreg32

	sw_flow_control portreg32

	flow_control_config reg32

	mac_rsv_mask portreg32

	_ reg32

	mode reg32

	port_enable reg32

	soft_reset reg32

	power_save reg32

	_ [0x210 - 0x20e]reg32

	mac_control reg32

	counter_mode reg32

	_ reg32

	phy_pll_status reg32

	phy_control reg32

	phy_lane_status [4]reg32

	phy_uc_data_access_mode reg32

	_ [0x224 - 0x21a]reg32

	reset_mib_counters reg32

	time_stamp_timer [2]reg32

	link_status_down reg32

	link_status_down_clear reg32

	interrupt_status reg32
	interrupt_enable reg32

	sbus_control reg32

	_ [0x600 - 0x22c]reg32
}

type xclmac_common_regs_0 struct {
	control portreg64

	mode portreg64

	spare [2]portreg64

	tx_control portreg64

	tx_src_address portreg64

	rx struct {
		control portreg64

		src_address portreg64

		max_bytes_per_packet portreg64

		ethernet_type_for_vlan portreg64

		lss_control portreg64

		lss_status       portreg64
		clear_lss_status portreg64
	}

	pause_control portreg64

	pfc struct {
		control     portreg64
		pfc_type    portreg64
		opcode      portreg64
		dst_address portreg64
	}

	llfc struct {
		control       portreg64
		tx_msg_fields portreg64
		rx_msg_fields portreg64
	}

	tx_timestamp_fifo_data   portreg64
	tx_timestamp_fifo_status portreg64

	fifo_status       portreg64
	fifo_status_clear portreg64

	lag_failover_status portreg64

	eee_control                 portreg64
	eee_timers                  portreg64
	eee_1_sec_link_status_timer portreg64

	higig_hdr [2]portreg64

	gmii_eee_control       portreg64
	tx_timestamp_adjust    portreg64
	tx_corrupt_crc_control portreg64

	e2e struct {
		control       portreg64
		cc_module_hdr [2]portreg64
		cc_data_hdr   [2]portreg64
		fc_module_hdr [2]portreg64
		fc_data_hdr   [2]portreg64
	}

	tx_fifo_cell_count portreg64

	tx_fifo_cell_request_count portreg64

	memory_control               portreg64
	ecc_control                  portreg64
	ecc_force_multiple_bit_error portreg64
	ecc_force_single_bit_error   portreg64
}

type xclmac_common_regs_1 struct {
	rx_cdc_memory_ecc_status portreg64
	tx_cdc_memory_ecc_status portreg64
	ecc_status_clear         portreg64
	version                  portreg64
}

type clmac_uncommon_regs_0 struct {
	rx_timestamp_memory_ecc_status portreg64
}

type clport_regs struct {
	xclport_regs
	xclmac_common_regs_0
	clmac_uncommon_regs_0
	xclmac_common_regs_1
}

type xlport_regs struct {
	xclport_regs
	xclmac_common_regs_0
	xclmac_common_regs_1
}

func (p *PortBlock) get_regs() (*xclport_regs, *xclmac_common_regs_0, *xclmac_common_regs_1, *clmac_uncommon_regs_0) {
	if p.IsXlPort {
		x := (*xlport_regs)(m.BasePointer)
		return &x.xclport_regs, &x.xclmac_common_regs_0, &x.xclmac_common_regs_1, (*clmac_uncommon_regs_0)(nil)
	} else {
		x := (*clport_regs)(m.BasePointer)
		return &x.xclport_regs, &x.xclmac_common_regs_0, &x.xclmac_common_regs_1, &x.clmac_uncommon_regs_0
	}
}

type wc_ucmem_data_elt [4]uint32

func get_xclport_mems() *xclport_mems { return (*xclport_mems)(m.BasePointer) }

type xclport_mems struct {
	wc_ucmem_data m.Mem
}
