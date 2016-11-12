// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"fmt"
	"math"
)

type tdm_u32 struct {
	// Rx pipe versions are not "port regs" and so have GenReg bit explicitly set.
	Array [1 << m.Log2NPorts]m.Pu32
}

func tdmAccessType(b sbus.Block, pipe_index uint) (a sbus.Address, c sbus.AccessType, ri uint) {
	switch b {
	case BlockRxPipe:
		// For IPIPE pipe is encoded in sbus access type.
		c = sbus.Unique(pipe_index)
		ri = 0
		// Need to set GenReg bit but only for rx pipe.
		a = sbus.GenReg
	case BlockMmuSlice:
		// For MMU pipe is encoded in register index.  Also need to set base type.
		c = sbus.Single
		ri = pipe_index
		a = sbus.Address(mmuBaseTypeTxPipe)
	default:
		panic(fmt.Errorf("unexpected block %s", sbusBlockString(b)))
	}
	return
}

func (r *tdm_u32) get(q *DmaRequest, b sbus.Block, pipe_index uint, v *uint32) {
	a, c, ri := tdmAccessType(b, pipe_index)
	r.Array[ri].Get(&q.DmaRequest, a, b, c, v)
}
func (r *tdm_u32) set(q *DmaRequest, b sbus.Block, pipe_index uint, v uint32) {
	a, c, ri := tdmAccessType(b, pipe_index)
	r.Array[ri].Set(&q.DmaRequest, a, b, c, v)
}

// TDM controller shared between rx pipe and mmu interface to tx pipe.
type tdm_controller struct {
	_ [0x1]tdm_u32

	config tdm_u32

	high_speed_port_bitmap tdm_u32

	opportunistic_scheduler_config tdm_u32

	cpu_loopback_opportunistic_scheduler_config tdm_u32

	_ [0x8 - 0x5]tdm_u32

	over_subscription_group_config [8]tdm_u32

	_ [0x14 - 0x10]tdm_u32

	over_subscription_group_members [8][12]tdm_u32

	port_block_calendar [8][7]tdm_u32

	calendar_ecc_enable tdm_u32

	_ [0x4b2 - 0x4ad]tdm_u32

	dft_input tdm_u32

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
	port   rx_pipe_mmu_port_number
	phy_id uint8
}

func (e *tdm_calendar_entry) getSet(b []uint32, lo int, isSet bool) int {
	v := uint8(e.port)
	i := m.MemGetSetUint8((*uint8)(&v), b, lo+5, lo, isSet)
	i = m.MemGetSetUint8(&e.phy_id, b, i+3, i, isSet)
	if !isSet {
		e.port = rx_pipe_mmu_port_number(v)
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
	members []rx_pipe_mmu_port_number
}

func (g *tdm_over_subscription_group) reset() {
	if g.members != nil {
		g.members = g.members[:0]
	}
}

func (g *tdm_over_subscription_group) add_member(p rx_pipe_mmu_port_number) {
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
		e := tdm_calendar_entry{port: rx_pipe_mmu_port_invalid, phy_id: 0xf}
		if i < nTokens {
			e.port = rx_pipe_mmu_port_over_subscription
		}
		c.entryBuf[rx][i] = e
		c.entryBuf[tx][i] = e
	}
	c.entries[rx] = c.entryBuf[rx][:nTokens]
	c.entries[tx] = c.entryBuf[tx][:nTokens]
}

func (t *fe1a) compute_tdm_calendar(pipe uint) {
	const (
		bandwidthPerToken  = 2.5e9 // bits per second
		minPacketBytes     = 64    // bytes
		preambleBytes      = 8
		interFrameGapBytes = 12
	)

	p := &t.pipes[pipe].tdm_pipe
	c := &p.calendar

	coreFreq := t.GetSwitchCommon().CoreFrequencyInHz
	minSizedPacketsPerSec := bandwidthPerToken / (8 * (minPacketBytes + preambleBytes + interFrameGapBytes))
	nTokens := int(math.Floor(coreFreq / minSizedPacketsPerSec))
	if nTokens > len(c.entryBuf[rx]) {
		panic("not enough entries")
	}

	p.initEntries(nTokens)

	var fixedSlots [n_rx_tx][]rx_pipe_mmu_port_number
	// 10G of bandwidth for cpu/management ports on pipes 0-2
	// 5G of bandwidth for loopback port.
	// Misc slots: 5G internal idle1 (refresh), 2.5G idle2, 2.5G null
	if pipe == 3 {
		fixedSlots[rx] = []rx_pipe_mmu_port_number{
			rx_pipe_mmu_port_loopback,
			rx_pipe_mmu_port_idle2,
			rx_pipe_mmu_port_null,
			rx_pipe_mmu_port_loopback,
			rx_pipe_mmu_port_idle1,
			rx_pipe_mmu_port_idle2,
		}
		fixedSlots[tx] = []rx_pipe_mmu_port_number{
			rx_pipe_mmu_port_loopback,
			rx_pipe_mmu_port_null,
			rx_pipe_mmu_port_null,
			rx_pipe_mmu_port_loopback,
			rx_pipe_mmu_port_null,
			rx_pipe_mmu_port_idle2,
		}
	} else {
		fixedSlots[rx] = []rx_pipe_mmu_port_number{
			rx_pipe_mmu_port_any_pipe_cpu_or_management,
			rx_pipe_mmu_port_loopback,
			rx_pipe_mmu_port_any_pipe_cpu_or_management,
			rx_pipe_mmu_port_idle2,
			rx_pipe_mmu_port_null,
			rx_pipe_mmu_port_loopback,
			rx_pipe_mmu_port_any_pipe_cpu_or_management,
			rx_pipe_mmu_port_idle1,
			rx_pipe_mmu_port_any_pipe_cpu_or_management,
			rx_pipe_mmu_port_idle2,
		}
		fixedSlots[tx] = []rx_pipe_mmu_port_number{
			rx_pipe_mmu_port_any_pipe_cpu_or_management,
			rx_pipe_mmu_port_loopback,
			rx_pipe_mmu_port_any_pipe_cpu_or_management,
			rx_pipe_mmu_port_null,
			rx_pipe_mmu_port_null,
			rx_pipe_mmu_port_loopback,
			rx_pipe_mmu_port_any_pipe_cpu_or_management,
			rx_pipe_mmu_port_null,
			rx_pipe_mmu_port_any_pipe_cpu_or_management,
			rx_pipe_mmu_port_idle2,
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

func (t *fe1a) install_tdm_calendar(pipe uint) {
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
		t.mmu_slice_mems.tdm_calendar[ci].entries[pipe][i2].seta(q, &d[tx], BlockMmuSlice, sbus.Single, mmuBaseTypeTxPipe)
	}
	q.Do()

	// Configure and enable calendar.
	v := uint32(1 << 17) // enable bit
	// Set tdm calender index and length.
	v |= uint32(ci) << 16
	v |= uint32(len(c.entries[rx])-1) << (8 * ci)
	t.rx_pipe_controller.rx_buffer_tdm_scheduler.config.set(q, BlockRxPipe, pipe, v)
	t.mmu_slice_controller.tdm.config.set(q, BlockMmuSlice, pipe, v)
	q.Do()
}

func (t *fe1a) install_over_subscription_groups() {
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
		pi := rx_pipe_mmu_port_data_0 + rx_pipe_mmu_port_number(pb.GetPortIndex(p)&0x1f)
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
			t.rx_pipe_controller.rx_buffer_tdm_scheduler.over_subscription_group_config[gi].set(q, BlockRxPipe, pipe, v[0])
			t.mmu_slice_controller.tdm.over_subscription_group_config[gi].set(q, BlockMmuSlice, pipe, v[1])
			q.Do()

			// Set group members.
			for i := uint(0); i < l; i++ {
				var (
					pi     [2]rx_pipe_mmu_port_number
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

				t.rx_pipe_controller.rx_buffer_tdm_scheduler.over_subscription_group_members[gi][i].set(q, BlockRxPipe, pipe, w[0])
				t.mmu_slice_controller.tdm.over_subscription_group_members[gi][i].set(q, BlockMmuSlice, pipe, w[1])
			}
			// Invalidate unused group members.
			for i := l; i < uint(len(t.rx_pipe_controller.rx_buffer_tdm_scheduler.over_subscription_group_members)); i++ {
				w := uint32(0x3f | (0x7 << 6))
				t.rx_pipe_controller.rx_buffer_tdm_scheduler.over_subscription_group_members[gi][i].set(q, BlockRxPipe, pipe, w)
				t.mmu_slice_controller.tdm.over_subscription_group_members[gi][i].set(q, BlockMmuSlice, pipe, w)
			}

			// Advance to next group.
			gi++
		}

		// Set high speed port bitmap for this pipe.
		t.rx_pipe_controller.rx_buffer_tdm_scheduler.high_speed_port_bitmap.set(q, BlockRxPipe, pipe, high_speed_port_bitmap[pipe][0])
		t.mmu_slice_controller.tdm.high_speed_port_bitmap.set(q, BlockMmuSlice, pipe, high_speed_port_bitmap[pipe][1])
	}

	q.Do()
}

func (t *fe1a) install_port_block_calendar() {
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
			rx_pipe_port_num := rx_pipe_mmu_port_number(0x3f)
			mmu_port_num := rx_pipe_port_num
			sub_pbi := port_block_index % 8 // block index within pipe
			if sub_port_index := int(pb.tdm_calendar[ci]); sub_port_index != -1 {
				// Correct spacing for lower speed mmu port calendar.
				spacing[0], spacing[1] = 4, 4
				if pb.speeds[sub_port_index] < 40e9 {
					spacing[1] = 8
				}
				v[0], v[1] = valid, valid
				rx_pipe_port_num = rx_pipe_mmu_port_number(sub_port_index + sub_pbi*4)
				mmu_port_num = rx_pipe_port_num.gratuitous_mmu_port_scrambling()
				v[0] |= spacing[0] << 7
				v[1] |= spacing[1] << 7
				v[0] |= uint32(rx_pipe_port_num) << 1
				v[1] |= uint32(mmu_port_num) << 1
			} else {
				v[0], v[1] = 0, 0
			}
			t.rx_pipe_controller.rx_buffer_tdm_scheduler.port_block_calendar[sub_pbi][ci].set(q, BlockRxPipe, pb.pipe, v[0])
			t.mmu_slice_controller.tdm.port_block_calendar[sub_pbi][ci].set(q, BlockMmuSlice, pb.pipe, v[1])
		}
	}
	q.Do()
}

func (t *fe1a) opportunistic_scheduler_init(pipe uint) {
	q := t.getDmaReq()
	var (
		v [2]uint32
		u [2]uint32
	)
	t.rx_pipe_controller.rx_buffer_tdm_scheduler.opportunistic_scheduler_config.get(q, BlockRxPipe, pipe, &v[0])
	t.mmu_slice_controller.tdm.opportunistic_scheduler_config.get(q, BlockMmuSlice, pipe, &v[1])
	t.rx_pipe_controller.rx_buffer_tdm_scheduler.cpu_loopback_opportunistic_scheduler_config.get(q, BlockRxPipe, pipe, &u[0])
	t.mmu_slice_controller.tdm.cpu_loopback_opportunistic_scheduler_config.get(q, BlockMmuSlice, pipe, &u[1])
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
		v[i] = (v[i] &^ (0x3f << 14)) | rx_pipe_mmu_port_null<<14
		u[i] |= enable_cpu_port | enable_loopback_port
	}
	t.rx_pipe_controller.rx_buffer_tdm_scheduler.opportunistic_scheduler_config.set(q, BlockRxPipe, pipe, v[0])
	t.mmu_slice_controller.tdm.opportunistic_scheduler_config.set(q, BlockMmuSlice, pipe, v[1])
	t.rx_pipe_controller.rx_buffer_tdm_scheduler.cpu_loopback_opportunistic_scheduler_config.set(q, BlockRxPipe, pipe, u[0])
	t.mmu_slice_controller.tdm.cpu_loopback_opportunistic_scheduler_config.set(q, BlockMmuSlice, pipe, u[1])
	q.Do()
}

func (t *fe1a) tdm_scheduler_init() {
	for pipe := uint(0); pipe < n_pipe; pipe++ {
		t.compute_tdm_calendar(pipe)
		t.install_tdm_calendar(pipe)
		t.opportunistic_scheduler_init(pipe)
	}
	t.install_over_subscription_groups()
	t.install_port_block_calendar()
}
