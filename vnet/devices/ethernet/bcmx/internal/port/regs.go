// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package port

import (
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"
)

type reg32 m.Reg32

func (r *reg32) get(q *dmaRequest, v *uint32) {
	(*m.Reg32)(r).Get(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}
func (r *reg32) getDo(q *dmaRequest) (v uint32) { r.get(q, &v); q.Do(); return }

func (r *reg32) set(q *dmaRequest, v uint32) {
	(*m.Reg32)(r).Set(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}

func (r *reg32) offset() uint          { return (*m.Reg32)(r).Offset() }
func (r *reg32) address() sbus.Address { return (*m.Reg32)(r).Address() }

type preg32 m.Preg32
type portreg32 [1 << m.Log2NRegPorts]preg32
type preg64 m.Preg64
type portreg64 [1 << m.Log2NRegPorts]preg64

func (r *preg32) get(q *dmaRequest, v *uint32) {
	(*m.Preg32)(r).Get(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}
func (r *preg32) getDo(q *dmaRequest) (v uint32) { r.get(q, &v); q.Do(); return }

func (r *preg32) set(q *dmaRequest, v uint32) {
	(*m.Preg32)(r).Set(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}

func (r *preg64) get(q *dmaRequest, v *uint64) {
	(*m.Preg64)(r).Get(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}
func (r *preg64) getDo(q *dmaRequest) (v uint64) { r.get(q, &v); q.Do(); return }

func (r *preg64) set(q *dmaRequest, v uint64) {
	(*m.Preg64)(r).Set(&q.DmaRequest, 0, q.portBlock.SbusBlock, sbus.Unique0, v)
}

func (r *preg32) offset() uint          { return (*m.Preg32)(r).Offset() }
func (r *preg32) address() sbus.Address { return (*m.Preg32)(r).Address() }
func (r *preg64) offset() uint          { return (*m.Preg64)(r).Offset() }
func (r *preg64) address() sbus.Address { return (*m.Preg64)(r).Address() }

// Registers shared between xlport & clport.
type xclport_regs struct {
	// 40/48 bit port packet/byte counters.  See counter_type for types.
	// 48 bit byte counters over flow in ~ 6 hours
	counters [N_counters]portreg64
	_        [0x200 - N_counters]portreg64

	// [11] hi gig 2 mode
	// [10] hi gig mode (deprecated)
	// [9:1]] hi gig my module id
	higig_config portreg32

	// The max packet size for counter updates.  Default is 1518.
	mib_stats_max_packet_size portreg32

	// [0] link status up
	lag_failover_config portreg32

	// [0] 0 => rx/tx eee counters asymmetric mode; 1 => symmetric mode
	eee_counter_mode portreg32
	_                reg32

	// [1] local fault implies link down to cmic
	// [0] remote fault implies link down to cmic
	link_status_to_cmic_control portreg32

	// [1] enable
	// [0] transition 0 -> 1 sends xoff packet out port.
	//     transition 1 -> 0 sends xon packet out port.
	sw_flow_control portreg32

	// [1] merge mode enable
	// [0] parallel flow control enable
	flow_control_config reg32

	// Control which RSV event cause "purge" event that triggers RXERR to be set for the packet by the MAC.
	mac_rsv_mask portreg32

	_ reg32

	/* [10] single port mode speed is 100g/120g
	   [9] use cmic time stamp instead of clport (default: 1)
	   [5:3] core port mode
	     0 => 4 ports: lane 0 1 2 3 single
	     1 => 3 ports: lane 2/3 dual 0/1 single
	     2 => 3 ports: lane 0/1 dual 2/3 single
	     3 => 2 ports: lane 0/1 dual lane 2/3 dual
	     4 => 1 port:  lanes 0/1/2/3
	   [2:0] phy mode (as above) only applies to xlport. */
	mode reg32

	/* [3:0] sub ports to enable. */
	port_enable reg32

	/* [3:0] active high port soft reset. */
	soft_reset reg32

	/* [0] enable clock gating for CLPORT core. */
	power_save reg32

	_ [0x210 - 0x20e]reg32

	/* [3] wan ipg enable
	   [2] flexible tdm
	   [1] bypass os/ts path for less latency
	   [0] active high reset (default value: 1) */
	mac_control reg32

	/* [0] 0 => saturate; clear on read, 1 => wrap on overflow.  Default is wrap. */
	counter_mode reg32

	_ reg32

	/* [1] sticky 1->0 loss of pll lock
	   [0] live pll lock */
	tsc_pll_status reg32

	/* [4] stop analog clocks but leave power to analog section
	   [3] tsc power down analog section (default value: 1)
	   [2] pad ( = 0) versus lcref (= 1) as ref clock source for tsc pll (default: 1)
	   [1] ref out enable for this tsc
	   [0] active low hard reset for this TSC */
	tsc_control reg32

	/* [2] pmd lock
	   [1] signal detect
	   [0] link status */
	tsc_lane_status [4]reg32

	/* [0] 0 => access tsc registers, 1 => tsc uc memory. */
	tsc_uc_data_access_mode reg32

	_ [0x224 - 0x21a]reg32

	/* [3:0] bit i for port i; 1 => reset; 0 => not reset */
	reset_mib_counters reg32

	time_stamp_timer [2]reg32

	/* [3:0] port mask sticky link status 1 -> 0 */
	link_status_down reg32

	/* [3:0] port mask */
	link_status_down_clear reg32

	// 8:5 RX_FLOWCONTROL_REQ_FULL
	// 4:4 TSC_ERR interrupt enable
	// 3:3 MAC_RX_CDC_MEM_ERR
	// 2:2 MAC_TX_CDC_MEM_ERR
	// 1:1 MIB_RX_MEM_ERR
	// 0:0 MIB_TX_MEM_ERR
	interrupt_status reg32
	interrupt_enable reg32

	sbus_control reg32

	_ [0x600 - 0x22c]reg32
}

// Registers shared between xl & cl macs group 0.
type xclmac_common_regs_0 struct {
	// 	[15] extended hg2 header enable (default 1)
	// 	[14] allow 40byte and greater packets
	// 	[13] link status from strap pin (versus from software)
	// 	[12] Link status indication from Software. If set, indicates that link 0x1
	//       is active. When this transitions from 0 to 1, EEE FSM waits
	//       for 1 second before it starts its operation (default 1)
	// 	[11] xgmii ipg check disable
	// 	[10] rs layer reset
	// 	[8] local loopback leak enable
	// 	[7] xlgmii align enable
	// 	[6] mac soft reset (default 1)
	// 	[5] LAG failover loopback
	// 	[4] remove LAG failover loopback
	// 	[2] local loopback enable
	// 	[1] rx enable
	// 	[0] tx enable
	control portreg64

	/* [6:4] speed (CLMAC: 2 => 1G, 4 => 100G; XLMAC: 10M 100M 1G 2.5G 10G)
	   [3] exclude SOP byte from crc calculation in hi-gig modes
	   [2:0] packet header mode (0 IEEE, 1 higig, 2 higig2). */
	mode portreg64

	spare [2]portreg64

	/* [41:38] tx threshold: number of 48 byte cells that are buffered in CDC fifo per packet (default: 1)
	   [37] egress pipe discard (don't write to CDC fifo)
	   [36:33] tx preamble length
	   [32:25] number of bytes to transmit before adding throt_num bytes to IPG (default: 8)
	   [24:19] throt_num number of extra ipg bytes
	   [18:12] average ipg
	   [11:5] min packet size (smaller packets are padded to this size)
	   [4] pad enable
	   [3] don't force first byte of packet to be START
	   [2] accept packets from host but do not transmit
	   [1:0] crc mode (0 append, 1 keep, 2 replace (default), 3 per packet). */
	tx_control portreg64

	// Used for pause frames.
	tx_src_address portreg64

	rx struct {
		/* [12] process variable preamble
		   [10:4] runt threshold: smaller packets are dropped.
		   [3] strict preamble
		   [2] strip crc
		   [1] any non idle character starts packet (not necessarily START byte) */
		control portreg64

		// Source address in addition to 01:80:c2:00:00:01 for control frames.
		src_address portreg64

		/* Default 1518 bytes.  Cannot be modified with traffic running. */
		max_bytes_per_packet portreg64

		/* [33] outer vlan enable
		   [32] inner vlan enable
		   [31:16] outer vlan ethernet type (0x8100 default)
		   [15:0] inner vlan ethernet type. */
		ethernet_type_for_vlan portreg64

		// [0] local fault disable
		// [1] remote fault disable
		// [2] use external faults for tx
		// [3] link interruption disable
		// [4] drop tx data on local fault (default 1)
		// [5] drop tx data on remote fault (default 1)
		// [6] drop tx data on link interrupt
		// [7] reset flow control timers on link down
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

	// [1] rx msg overflow
	// [2] tx packet underflow
	// [3] tx packet overflow
	// [5] tx llfc msg overflow
	// [6] tx timestamp fifo overflow
	// [7] rx packet overflow
	// [8] link status
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

	// [5:0] cell count
	tx_fifo_cell_count portreg64
	// [5:0] cell request count
	tx_fifo_cell_request_count portreg64

	memory_control               portreg64
	ecc_control                  portreg64
	ecc_force_multiple_bit_error portreg64
	ecc_force_single_bit_error   portreg64
}

// Registers shared between xl & cl macs group 1.
type xclmac_common_regs_1 struct {
	rx_cdc_memory_ecc_status portreg64
	tx_cdc_memory_ecc_status portreg64
	ecc_status_clear         portreg64
	version                  portreg64
}

type clmac_uncommon_regs_0 struct {
	// not present in xlmac; everything else is the same.
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
		x := (*xlport_regs)(m.RegsBasePointer)
		return &x.xclport_regs, &x.xclmac_common_regs_0, &x.xclmac_common_regs_1, (*clmac_uncommon_regs_0)(nil)
	} else {
		x := (*clport_regs)(m.RegsBasePointer)
		return &x.xclport_regs, &x.xclmac_common_regs_0, &x.xclmac_common_regs_1, &x.clmac_uncommon_regs_0
	}
}

type wc_ucmem_data_elt [4]uint32

func get_xclport_mems() *xclport_mems { return (*xclport_mems)(m.RegsBasePointer) }

// Port memories.
type xclport_mems struct {
	wc_ucmem_data m.Mem
}
