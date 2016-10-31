// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/sbus"

	"fmt"
	"math"
)

type tdm_reg32 struct {
	// Ipipe versions are not "port regs" and so have GenReg bit explicitly set.
	Array [1 << m.Log2NRegPorts]m.Preg32
}

func tdmAccessType(b sbus.Block, pipe_index uint) (a sbus.Address, c sbus.AccessType, ri uint) {
	switch b {
	case BlockRxPipe:
		// For IPIPE pipe is encoded in sbus access type.
		c = sbus.Unique(pipe_index)
		ri = 0
		// Need to set GenReg bit but only for rx pipe.
		a = sbus.GenReg
	case BlockMmuSc:
		// For MMU pipe is encoded in register index.  Also need to set base type.
		c = sbus.Single
		ri = pipe_index
		a = sbus.Address(mmuBaseTypeTxPipe)
	default:
		panic(fmt.Errorf("unexpected block %s", sbusBlockString(b)))
	}
	return
}

func (r *tdm_reg32) get(q *DmaRequest, b sbus.Block, pipe_index uint, v *uint32) {
	a, c, ri := tdmAccessType(b, pipe_index)
	r.Array[ri].Get(&q.DmaRequest, a, b, c, v)
}
func (r *tdm_reg32) set(q *DmaRequest, b sbus.Block, pipe_index uint, v uint32) {
	a, c, ri := tdmAccessType(b, pipe_index)
	r.Array[ri].Set(&q.DmaRequest, a, b, c, v)
}

// TDM registers shared between rx pipe and mmu.
type tdm_regs struct {
	_ [0x1]tdm_reg32

	// Address: 0x04040100 Block ID: IPIPE Access Type: UNIQUE_PIPE0123
	// Description: TDM calendar config register
	// 17:17 ENABLE R/W Enables the TDM port pick function
	// 16:16 CURR_CAL R/W Indicates which calendar is to be used currently
	// 15:8 CAL1_END R/W TDM Calendar1 end entry
	// 7:0 CAL0_END R/W TDM Calendar0 end entry
	config tdm_reg32

	// Address: 0x04040200 Block ID: IPIPE Access Type: UNIQUE_PIPE0123
	// Description: TDM_HSP HIGH SPEED PORT indication (high speed = speed >= 40e9 bits/sec)
	// 31:0 PORT_BMP R/W Setting this bit indicates that this port is a high speed port requiring minimum spacing of less than 8
	//   (minimum spacing is fixed for Loopback at 4 and CPU at 8)
	high_speed_port_bitmap tdm_reg32

	// Address: 0x04040300 Block ID: IPIPE Access Type: UNIQUE_PIPE0123
	// Description: Post Calendar opportunistic scheduler config
	// 19:14 DISABLE_PORT_NUM R/W 0x25 If this port number is present in the Calendar then it cannot be substituted by opportunistic scheduler
	// 13:11 OPP_STRICT_PRI R/W Control strict priority scheduling between 3 opportunistic sources (high to low)
	//   0: Opp1, Opp2, Oversub
	//   1: Opp2, Opp1, Oversub
	//   2: Opp1, Oversub, Opp2
	//   3: Opp2, Oversub, Opp1
	//   4: Oversub, Opp1, Opp2
	//   5: Oversub, Opp2, Opp1
	// 10:10 OPP_OVR_SUB_EN R/W Enable for oversub port in opportunistic scheduling
	// 9:6 OPP2_SPACING R/W Same spacing value for OPP2_PORT_NUM
	// 5:5 OPP2_PORT_EN R/W Enable for opportunistic2 port in opportunistic scheduling
	// 4:1 OPP1_SPACING R/W Same spacing value for OPP1_PORT_NUM
	// 0:0 OPP1_PORT_EN R/W Enable for opportunistic1 port in opportunistic scheduling
	opportunistic_scheduler_config tdm_reg32

	// Address: 0x04040400 Block ID: IPIPE Access Type: UNIQUE_PIPE0123
	// Description: CPU Loopback opportunistic scheduler config
	// 13:10 LB_SPACING R/W 0x4 Minimum spacing for Loopback picks
	// 9:6 CPU_SPACING R/W 0x8 Minimum spacing for CPU picks
	// 5:2 LB_CPU_RATIO R/W 0xa Ratio of number of opportunistic slots to be given to Loopback to CPU
	// 1:1 LB_OPP_EN R/W Enable for Loopback port in opportunistic scheduling
	// 0:0 CPU_OPP_EN R/W Enable for CPU port in opportunistic scheduling
	cpu_loopback_opportunistic_scheduler_config tdm_reg32

	_ [0x8 - 0x5]tdm_reg32

	// Number of Register Instances: 8 groups
	// Address: 0x04040800 Block ID: IPIPE Access Type: UNIQUE_PIPE0123
	// Description: TDM Oversub group configuration registers
	// 9:7 SPEED R/W Speed of this group. The only values allowed are 2,4,5,8,10 and 20
	//   0x0 = 0
	//   0x1 = 2 < 20G
	//   0x2 = 4 >= 20G
	//   0x3 = 5 >= 25G
	//   0x4 = 8 >= 40G
	//   0x5 = 10 >= 50G
	//   0x6 = 20 >= 100G
	// 6:4 SISTER_SPACING R/W 0x4 spacing between ports within same PHY_PORT_ID
	// 3:0 SAME_SPACING R/W 0x8 spacing between same ports (for mmu: set to >= 40G = 4; else 8; for idb: 4)
	over_subscription_group_config [8]tdm_reg32

	_ [0x14 - 0x10]tdm_reg32

	// Number of Register Instances: 8 groups x 12 max members/group
	// Address: 0x04041400 + 0xc00*i i=0..7 Block ID: IPIPE Access Type: UNIQUE_PIPE0123
	// Description: TDM Oversub group0 Member Table
	// Index Description: oversub group table entry #
	// 8:6 PHY_PORT_ID Physical Port ID. All ports which belong to the same serdes are configured to the same PHY_PORT_ID. 0x7
	// 5:0 PORT_NUM R/W MMU Port number residing in the current entry. Unused entries are configured with reserved MMU port number 0x3f
	// All group member ports have the same speed.
	over_subscription_group_members [8][12]tdm_reg32

	// Number of Register Instances: 8 port-blocks-per-pipe x 7 calendar entry time-slots
	// Address: 0x04047400 + i*0x700 i = 0..7 Block ID: IPIPE Access Type: UNIQUE_PIPE0123
	// Description: Port block Calendar
	// Index Description: PBLK0_CALENDAR entry #
	// 10:7 SPACING R/W The same spacing value for this port.  IDB => always 4; MMU 4 for >= 40G else 8
	// 6:1 PORT_NUM R/W 0x3f IDB port number residing in the current entry. Unused entries are configured with reserved port number 0x3f
	// 0:0 VALID R/W Indicates this entry is valid
	port_block_calendar [8][7]tdm_reg32

	// Address: 0x0404ac00 Block ID: IPIPE Access Type: UNIQUE_PIPE0123
	// Description: MMU TDM 1bit correctable memory ECC error reporting enable
	// 0:0 TDM_CAL R/W Enable TDM0_CAL 1bit ECC reporting
	calendar_ecc_enable tdm_reg32

	_ [0x4b2 - 0x4ad]tdm_reg32

	// Address: 0x0404b200 Block ID: IPIPE Access Type: UNIQUE_PIPE0123
	// Description: DFT input
	// 11:0 CAL_TM R/W TM bits
	dft_input tdm_reg32

	_ [0x08000000 - 0x0404b300]byte
}

type tdm_calendar_mem m.MemElt

func (x *tdm_calendar_mem) geta(q *DmaRequest, e *tdm_calendar_dual_entry, b sbus.Block, t sbus.AccessType, a sbus.Address) {
	(*m.MemElt)(x).MemDmaGeta(&q.DmaRequest, e, b, t, a)
}
func (x *tdm_calendar_mem) seta(q *DmaRequest, e *tdm_calendar_dual_entry, b sbus.Block, t sbus.AccessType, a sbus.Address) {
	(*m.MemElt)(x).MemDmaSeta(&q.DmaRequest, e, b, t, a)
}

type tdm_calendar_entry struct {
	port   idb_mmu_port_number
	phy_id uint8
}

func (e *tdm_calendar_entry) getSet(b []uint32, lo int, isSet bool) int {
	v := uint8(e.port)
	i := m.MemGetSetUint8((*uint8)(&v), b, lo+5, lo, isSet)
	i = m.MemGetSetUint8(&e.phy_id, b, i+3, i, isSet)
	if !isSet {
		e.port = idb_mmu_port_number(v)
	}
	return i
}

type tdm_calendar_dual_entry [2]tdm_calendar_entry

func (r *tdm_calendar_dual_entry) MemBits() int { return 27 }
func (r *tdm_calendar_dual_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for j := range r {
		i = r[j].getSet(b, i, isSet)
	}
	if i != 20 {
		panic("tdm")
	}
}

type tdm_over_subscription_group struct {
	members []idb_mmu_port_number
}

func (g *tdm_over_subscription_group) reset() {
	if g.members != nil {
		g.members = g.members[:0]
	}
}

func (g *tdm_over_subscription_group) add_member(p idb_mmu_port_number) {
	g.members = append(g.members, p)
}
func (g *tdm_over_subscription_group) n_members() uint { return uint(len(g.members)) }

type tdm_pipe struct {
	calendar struct {
		entryBuf [n_rx_tx][256]tdm_calendar_entry
		entries  [n_rx_tx][]tdm_calendar_entry
		// Incremented once for each install.
		installCount uint64
	}
	over_subscription_groups [n_port_speed]tdm_over_subscription_group
}

func (p *tdm_pipe) initEntries(nTokens int) {
	c := &p.calendar
	for i := range c.entryBuf[rx] {
		e := tdm_calendar_entry{port: idb_mmu_port_invalid, phy_id: 0xf}
		if i < nTokens {
			e.port = idb_mmu_port_over_subscription
		}
		c.entryBuf[rx][i] = e
		c.entryBuf[tx][i] = e
	}
	c.entries[rx] = c.entryBuf[rx][:nTokens]
	c.entries[tx] = c.entryBuf[tx][:nTokens]
}

func (p *tdm_pipe) initSdk(pipe uint) {
	nTokens := 215
	p.initEntries(nTokens)
	c := &p.calendar

	if pipe == 3 {
		i := 41
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_loopback, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_loopback, phy_id: 0xf}
		i += 1 + 41
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_idle2, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_null, phy_id: 0xf}
		i += 1 + 20
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_null, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_null, phy_id: 0xf}
		i += 1 + 21
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_loopback, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_loopback, phy_id: 0xf}
		i += 1 + 43
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_idle1, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_null, phy_id: 0xf}
		i += 1 + 43
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_idle2, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_idle2, phy_id: 0xf}
		i += 1
		if i != nTokens {
			panic("nTokens")
		}
	} else {
		i := 20
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_any_pipe_cpu_or_management, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_any_pipe_cpu_or_management, phy_id: 0xf}
		i += 1 + 20
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_loopback, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_loopback, phy_id: 0xf}
		i += 1 + 20
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_any_pipe_cpu_or_management, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_any_pipe_cpu_or_management, phy_id: 0xf}
		i += 1 + 20
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_idle2, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_null, phy_id: 0xf}
		i += 1 + 20
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_null, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_null, phy_id: 0xf}
		i += 1 + 21
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_loopback, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_loopback, phy_id: 0xf}
		i += 1 + 21
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_any_pipe_cpu_or_management, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_any_pipe_cpu_or_management, phy_id: 0xf}
		i += 1 + 21
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_idle1, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_null, phy_id: 0xf}
		i += 1 + 21
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_any_pipe_cpu_or_management, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_any_pipe_cpu_or_management, phy_id: 0xf}
		i += 1 + 21
		c.entries[rx][i] = tdm_calendar_entry{port: idb_mmu_port_idle2, phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: idb_mmu_port_idle2, phy_id: 0xf}
		i += 1
		if i != nTokens {
			panic("nTokens")
		}
	}
}

func (t *tomahawk) compute_tdm_calendar(pipe uint) {
	const (
		bandwidthPerToken  = 2.5e9 // bits per second
		minPacketBytes     = 64    // bytes
		preambleBytes      = 8
		interFrameGapBytes = 12
	)

	p := &t.pipes[pipe].tdm_pipe
	c := &p.calendar

	if false {
		p.initSdk(pipe)
		return
	}

	coreFreq := t.GetSwitchCommon().CoreFrequencyInHz
	minSizedPacketsPerSec := bandwidthPerToken / (8 * (minPacketBytes + preambleBytes + interFrameGapBytes))
	nTokens := int(math.Floor(coreFreq / minSizedPacketsPerSec))
	if nTokens > len(c.entryBuf[rx]) {
		panic("not enough entries")
	}

	p.initEntries(nTokens)

	var fixedSlots [n_rx_tx][]idb_mmu_port_number
	// 10G of bandwidth for cpu/management ports on pipes 0-2
	// 5G of bandwidth for loopback port.
	// Misc slots: 5G internal idle1 (refresh), 2.5G idle2, 2.5G null
	if pipe == 3 {
		fixedSlots[rx] = []idb_mmu_port_number{
			idb_mmu_port_loopback,
			idb_mmu_port_idle2,
			idb_mmu_port_null,
			idb_mmu_port_loopback,
			idb_mmu_port_idle1,
			idb_mmu_port_idle2,
		}
		fixedSlots[tx] = []idb_mmu_port_number{
			idb_mmu_port_loopback,
			idb_mmu_port_null,
			idb_mmu_port_null,
			idb_mmu_port_loopback,
			idb_mmu_port_null,
			idb_mmu_port_idle2,
		}
	} else {
		fixedSlots[rx] = []idb_mmu_port_number{
			idb_mmu_port_any_pipe_cpu_or_management,
			idb_mmu_port_loopback,
			idb_mmu_port_any_pipe_cpu_or_management,
			idb_mmu_port_idle2,
			idb_mmu_port_null,
			idb_mmu_port_loopback,
			idb_mmu_port_any_pipe_cpu_or_management,
			idb_mmu_port_idle1,
			idb_mmu_port_any_pipe_cpu_or_management,
			idb_mmu_port_idle2,
		}
		fixedSlots[tx] = []idb_mmu_port_number{
			idb_mmu_port_any_pipe_cpu_or_management,
			idb_mmu_port_loopback,
			idb_mmu_port_any_pipe_cpu_or_management,
			idb_mmu_port_null,
			idb_mmu_port_null,
			idb_mmu_port_loopback,
			idb_mmu_port_any_pipe_cpu_or_management,
			idb_mmu_port_null,
			idb_mmu_port_any_pipe_cpu_or_management,
			idb_mmu_port_idle2,
		}
	}

	nFixedSlots := len(fixedSlots[rx])
	div, rem := nTokens/nFixedSlots, nTokens%nFixedSlots
	r, dr := float64(0), float64(rem)/float64(nFixedSlots)
	i, slot := 0, 0
	for {
		c.entries[rx][i] = tdm_calendar_entry{port: fixedSlots[rx][slot], phy_id: 0xf}
		c.entries[tx][i] = tdm_calendar_entry{port: fixedSlots[tx][slot], phy_id: 0xf}
		slot++
		if slot >= nFixedSlots {
			break
		}
		i += div
		r += dr
		if r > 1 {
			i++
			r -= 1
		}
	}
}

func (t *tomahawk) install_tdm_calendar(pipe uint) {
	c := &t.pipes[pipe].tdm_pipe.calendar

	ci := uint32(c.installCount % 2)
	c.installCount++

	// DMA calendar to both ingress scheduler and mmu sc/tx pipe.
	q := t.getDmaReq()
	for i := 0; i < len(c.entryBuf[rx]); i += 2 {
		i2 := i / 2

		var d [n_rx_tx]tdm_calendar_dual_entry
		for j := range d {
			d[j][0] = c.entryBuf[j][i+0]
			d[j][1] = c.entryBuf[j][i+1]
		}
		t.rx_pipe_mems.tdm_calendar[ci].entries[i2].seta(q, &d[rx], BlockRxPipe, sbus.Unique(pipe), 0)
		t.mmu_sc_mems.tdm_calendar[ci].entries[pipe][i2].seta(q, &d[tx], BlockMmuSc, sbus.Single, mmuBaseTypeTxPipe)
	}
	q.Do()

	// Configure and enable calendar.
	v := uint32(1 << 17) // enable bit
	// Set tdm calender index and length.
	v |= uint32(ci) << 16
	v |= uint32(len(c.entries[rx])-1) << (8 * ci)
	t.rx_pipe_regs.rx_buffer_tdm_scheduler.config.set(q, BlockRxPipe, pipe, v)
	t.mmu_sc_regs.tdm.config.set(q, BlockMmuSc, pipe, v)
	q.Do()
}

func (t *tomahawk) install_over_subscription_groups() {
	q := t.getDmaReq()
	sc := t.GetSwitchCommon()

	var (
		high_speed_port_bitmap [n_pipe][2]uint32
	)

	// First empty over subscription groups.
	for i := range t.pipes {
		for j := range t.pipes[i].over_subscription_groups {
			t.pipes[i].over_subscription_groups[j].reset()
		}
	}

	// Compute over subscription group members based on port speeds.
	for _, p := range sc.Ports {
		pc := p.GetPortCommon()
		if pc.IsManagement || !m.IsProvisioned(p) {
			continue
		}
		pb := pc.PortBlock
		pipe := pb.GetPipe()
		sp := port_speed_code(pc.SpeedBitsPerSec)
		pi := idb_mmu_port_data_0 + idb_mmu_port_number(pb.GetPortIndex(p)&0x1f)
		t.pipes[pipe].over_subscription_groups[sp].add_member(pi)
		if pc.SpeedBitsPerSec >= 40e9 {
			mpi := pi.gratuitous_mmu_port_scrambling()
			high_speed_port_bitmap[pipe][0] |= 1 << pi
			high_speed_port_bitmap[pipe][1] |= 1 << mpi
		}
	}

	// Install in hardware.
	for pipe := uint(0); pipe < n_pipe; pipe++ {
		tp := &t.pipes[pipe].tdm_pipe
		gi := 0
		for speed_index := 0; speed_index < n_port_speed-1; speed_index++ { // only speeds 1-6 are used.
			speed_code := port_speed_first + speed_index
			l := tp.over_subscription_groups[speed_code].n_members()
			if l == 0 {
				continue
			}

			var v [2]uint32
			// Spacing of 4 for same phy id; spacing of 4 for same port.
			v[0] = (4 << 4) | (4 << 0)
			v[1] = v[0]
			// For mmu set spacing of 8 for lower speed ports (< 40G)
			if speed_code < port_speed_lt_40g {
				v[1] = (v[0] &^ 0xf) | (8 << 0)
			}
			w := uint32(speed_code) << 7
			v[0] |= w
			v[1] |= w
			t.rx_pipe_regs.rx_buffer_tdm_scheduler.over_subscription_group_config[gi].set(q, BlockRxPipe, pipe, v[0])
			t.mmu_sc_regs.tdm.over_subscription_group_config[gi].set(q, BlockMmuSc, pipe, v[1])
			q.Do()

			// Set group members.
			for i := uint(0); i < l; i++ {
				var (
					pi     [2]idb_mmu_port_number
					phy_id [2]uint32
					w      [2]uint32
				)

				// Be careful to mmu scramble mmu port index.
				pi[0] = tp.over_subscription_groups[speed_code].members[i]
				pi[1] = pi[0].gratuitous_mmu_port_scrambling()

				for j := range phy_id {
					phy_id[j] = uint32(((pi[0] / 4) & 0x7))
					w[j] = uint32(pi[j]) | phy_id[j]<<6
				}

				t.rx_pipe_regs.rx_buffer_tdm_scheduler.over_subscription_group_members[gi][i].set(q, BlockRxPipe, pipe, w[0])
				t.mmu_sc_regs.tdm.over_subscription_group_members[gi][i].set(q, BlockMmuSc, pipe, w[1])
			}
			// Invalidate unused group members.
			for i := l; i < uint(len(t.rx_pipe_regs.rx_buffer_tdm_scheduler.over_subscription_group_members)); i++ {
				w := uint32(0x3f | (0x7 << 6))
				t.rx_pipe_regs.rx_buffer_tdm_scheduler.over_subscription_group_members[gi][i].set(q, BlockRxPipe, pipe, w)
				t.mmu_sc_regs.tdm.over_subscription_group_members[gi][i].set(q, BlockMmuSc, pipe, w)
			}

			// Advance to next group.
			gi++
		}

		// Set high speed port bitmap for this pipe.
		t.rx_pipe_regs.rx_buffer_tdm_scheduler.high_speed_port_bitmap.set(q, BlockRxPipe, pipe, high_speed_port_bitmap[pipe][0])
		t.mmu_sc_regs.tdm.high_speed_port_bitmap.set(q, BlockMmuSc, pipe, high_speed_port_bitmap[pipe][1])
	}

	q.Do()
}

func (t *tomahawk) install_port_block_calendar() {
	sc := t.GetSwitchCommon()

	type portBlock struct {
		tdm_calendar [7]int8
		speeds       [4]float64
		pipe         uint
		n_ports      uint
	}
	var pbs [n_port_block_100g]portBlock

	for _, p := range sc.Ports {
		pc := p.GetPortCommon()
		if pc.IsManagement || !m.IsProvisioned(p) {
			continue
		}
		pb := pc.PortBlock
		bi := pb.GetPortBlockIndex()
		pi := pc.GetSubPortIndex()
		pbs[bi].pipe = pb.GetPipe()
		pbs[bi].speeds[pi] = pc.SpeedBitsPerSec
		pbs[bi].n_ports++
	}

	q := t.getDmaReq()
	for port_block_index := range pbs {
		pb := &pbs[port_block_index]
		switch pb.n_ports {
		case 1:
			// 1 port x 4 lanes all at the same speed
			pb.tdm_calendar = [7]int8{0, -1, 0, 0, -1, 0, -1}
		case 2:
			switch {
			case pb.speeds[2] >= 2*pb.speeds[0]:
				// port 0: 1S, port 2: 2S
				pb.tdm_calendar = [7]int8{0, 2, 2, 0, 2, 2, -1}
			case pb.speeds[0] >= 2*pb.speeds[2]:
				// port 0: 2S, port 2: 1S
				pb.tdm_calendar = [7]int8{2, 0, 0, 2, 0, 0, -1}
			default:
				// 2 ports at same speed.
				pb.tdm_calendar = [7]int8{0, -1, 2, 0, -1, 2, -1}
			}
		case 3:
			switch {
			case pb.speeds[2] == pb.speeds[3] && pb.speeds[0] >= 4*pb.speeds[2]:
				pb.tdm_calendar = [7]int8{0, 0, 2, 0, 0, 3, -1}
			case pb.speeds[2] == pb.speeds[3] && pb.speeds[0] >= 2*pb.speeds[2]:
				pb.tdm_calendar = [7]int8{0, -1, 2, 0, -1, 3, -1}
			case pb.speeds[0] == pb.speeds[1] && pb.speeds[2] >= 4*pb.speeds[0]:
				pb.tdm_calendar = [7]int8{0, 2, 2, 1, 2, 2, -1}
			case pb.speeds[0] == pb.speeds[1] && pb.speeds[2] >= 2*pb.speeds[0]:
				pb.tdm_calendar = [7]int8{0, -1, 2, 1, -1, 2, -1}
			default:
				panic("tri port")
			}
		case 4:
			pb.tdm_calendar = [7]int8{0, -1, 2, 1, -1, 3, -1}
		default:
			panic("n ports")
		}

		for ci := range pb.tdm_calendar {
			const (
				valid = 1 << 0
			)
			var (
				v       [2]uint32
				spacing [2]uint32
			)
			idb_port_num := idb_mmu_port_number(0x3f)
			mmu_port_num := idb_port_num
			sub_pbi := port_block_index % 8 // block index within pipe
			if sub_port_index := int(pb.tdm_calendar[ci]); sub_port_index != -1 {
				// Correct spacing for lower speed mmu port calendar.
				spacing[0], spacing[1] = 4, 4
				if pb.speeds[sub_port_index] < 40e9 {
					spacing[1] = 8
				}
				v[0], v[1] = valid, valid
				idb_port_num = idb_mmu_port_number(sub_port_index + sub_pbi*4)
				mmu_port_num = idb_port_num.gratuitous_mmu_port_scrambling()
				v[0] |= spacing[0] << 7
				v[1] |= spacing[1] << 7
				v[0] |= uint32(idb_port_num) << 1
				v[1] |= uint32(mmu_port_num) << 1
			} else {
				v[0], v[1] = 0, 0
			}
			t.rx_pipe_regs.rx_buffer_tdm_scheduler.port_block_calendar[sub_pbi][ci].set(q, BlockRxPipe, pb.pipe, v[0])
			t.mmu_sc_regs.tdm.port_block_calendar[sub_pbi][ci].set(q, BlockMmuSc, pb.pipe, v[1])
		}
	}
	q.Do()
}

func (t *tomahawk) opportunistic_scheduler_init(pipe uint) {
	q := t.getDmaReq()
	var (
		v [2]uint32
		u [2]uint32
	)
	t.rx_pipe_regs.rx_buffer_tdm_scheduler.opportunistic_scheduler_config.get(q, BlockRxPipe, pipe, &v[0])
	t.mmu_sc_regs.tdm.opportunistic_scheduler_config.get(q, BlockMmuSc, pipe, &v[1])
	t.rx_pipe_regs.rx_buffer_tdm_scheduler.cpu_loopback_opportunistic_scheduler_config.get(q, BlockRxPipe, pipe, &u[0])
	t.mmu_sc_regs.tdm.cpu_loopback_opportunistic_scheduler_config.get(q, BlockMmuSc, pipe, &u[1])
	q.Do()
	const (
		enable_opportunistic_port1    = 1 << 0
		enable_opportunistic_port2    = 1 << 5
		enable_over_subscription_port = 1 << 10
		enable_cpu_port               = 1 << 0
		enable_loopback_port          = 1 << 1
	)
	for i := range v {
		v[i] |= enable_opportunistic_port1 | enable_opportunistic_port2 | enable_over_subscription_port
		// Opportunistic disable port number: reset value is internal idle1; set to null port.
		v[i] = (v[i] &^ (0x3f << 14)) | idb_mmu_port_null<<14
		u[i] |= enable_cpu_port | enable_loopback_port
	}
	t.rx_pipe_regs.rx_buffer_tdm_scheduler.opportunistic_scheduler_config.set(q, BlockRxPipe, pipe, v[0])
	t.mmu_sc_regs.tdm.opportunistic_scheduler_config.set(q, BlockMmuSc, pipe, v[1])
	t.rx_pipe_regs.rx_buffer_tdm_scheduler.cpu_loopback_opportunistic_scheduler_config.set(q, BlockRxPipe, pipe, u[0])
	t.mmu_sc_regs.tdm.cpu_loopback_opportunistic_scheduler_config.set(q, BlockMmuSc, pipe, u[1])
	q.Do()
}

func (t *tomahawk) tdm_scheduler_init() {
	for pipe := uint(0); pipe < n_pipe; pipe++ {
		t.compute_tdm_calendar(pipe)
		t.install_tdm_calendar(pipe)
		t.opportunistic_scheduler_init(pipe)
	}
	t.install_over_subscription_groups()
	t.install_port_block_calendar()
}
