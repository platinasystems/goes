// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/elib/hw/pci"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"fmt"
	"time"
	"unsafe"
)

type top_u32 uint32

func (r *top_u32) set(q *DmaRequest, v uint32)  { q.SetU32(v, BlockTop, r.address(), sbus.Unique0) }
func (r *top_u32) get(q *DmaRequest, v *uint32) { q.GetU32(v, BlockTop, r.address(), sbus.Unique0) }
func (r *top_u32) getDo(q *DmaRequest) (v uint32) {
	r.get(q, &v)
	q.Do()
	return
}

func (r *top_u32) address() sbus.Address { return sbus.GenReg | sbus.Address(r.offset()) }
func (r *top_u32) offset() uint          { return uint(uintptr(unsafe.Pointer(r))-m.BaseAddress) << 8 }

type top_pll_controller struct {
	control [5]top_u32

	status top_u32
}

type top_controller struct {
	_ [0x2fc - 0]byte

	port_reset top_u32

	revision_id top_u32

	soft_reset [2]top_u32

	_ [0x380 - 0x30c]byte

	core_pll0_control    [5]top_u32
	core_pll1_control012 [3]top_u32

	_ [0x400 - 0x3a0]byte

	core_pll1_control34 [2]top_u32
	core_pll_status     [2]top_u32

	_ [0x420 - 0x410]byte

	xg_pll [4]top_pll_controller

	_ [0x498 - 0x480]byte

	ts_pll                         top_pll_controller
	bs_pll0                        top_pll_controller
	l1_received_clock_valid_status [3]top_u32
	bs_pll1                        top_pll_controller

	_ [0x500 - 0x4ec]byte

	temperature_sensor struct {
		control [2]top_u32
		current [9]top_u32
	}

	_ [0x600 - 0x52c]byte

	tsc_disable top_u32

	tsc_afe_pll_status top_u32

	_ [0x75c - 0x608]byte

	core_pll_frequency_select top_u32

	_ [0x77c - 0x760]byte

	l1_received_clock_valid_status34 top_u32
	misc_control_2                   top_u32
	freq_switch_status               top_u32

	port_enable            [4]top_u32
	management_port_enable top_u32
	tsc_disable_status     top_u32

	_ [0x800 - 0x7a0]byte

	temperature_sensor_interrupt struct {
		enable     top_u32
		thresholds [9]top_u32
		status     top_u32
	}

	_ [0x87c - 0x82c]byte

	tsc_resolved_speed_status [32]top_u32

	_ [0x914 - 0x8fc]byte

	avs_control top_u32
}

func (t *fe1a) setCoreFreq(q *DmaRequest) {
	f := uint32(0)
	switch t.CoreFrequencyInHz {
	case 950e6:
		f = 0
	case 850e6:
		f = 1
	case 765e6:
		f = 2
	case 672e6:
		f = 3
	case 645e6:
		f = 4
	case 545e6:
		f = 5
	default:
		panic(fmt.Errorf("unknown core frequency %eHz", t.CoreFrequencyInHz))
	}
	const core_clock0_sw_select_enable = 1 << 3
	v := t.top_controller.core_pll_frequency_select.getDo(q)
	v |= (f << 0) | core_clock0_sw_select_enable
	t.top_controller.core_pll_frequency_select.set(q, v)
	q.Do()
}

type pipe struct {
	tdm_pipe
}

type fe1a struct {
	m.SwitchCommon

	dmaReq DmaRequest

	revision_id revision_id_u32

	top_controller        *top_controller
	rx_pipe_mems          *rx_pipe_mems
	rx_pipe_controller    *rx_pipe_controller
	tx_pipe_mems          *tx_pipe_mems
	tx_pipe_controller    *tx_pipe_controller
	mmu_global_controller *mmu_global_controller
	mmu_global_mems       *mmu_global_mems
	mmu_pipe_controller   *mmu_pipe_controller
	mmu_pipe_mems         *mmu_pipe_mems
	mmu_slice_controller  *mmu_slice_controller
	mmu_slice_mems        *mmu_slice_mems

	mmu_pipe_by_phys_port [n_phys_ports]mmu_pipe

	pipes [n_pipe]pipe

	adjacency_main
	pipe_counter_main
	ip4_fib_main
	l3_main
	l3_interface_main
	l3_terminate_tcam_main
	portMain
	port_counter_main
	port_bitmap_main
	source_trunk_map_main
}

func (t *fe1a) getDmaReq() *DmaRequest { return &t.dmaReq }
func (t *fe1a) Interrupt()             { t.CpuMain.Interrupt() }
func (t *fe1a) GetLedDataRam()         { t.CpuMain.LedDataRamDump() }

func (t *fe1a) pll_wait(r *top_u32, tag string) {
	q := t.getDmaReq()
	start := time.Now()
	for {
		if s := r.getDo(q); s&(1<<31) != 0 {
			break
		} else {
			if time.Since(start) > 100*time.Millisecond {
				panic(fmt.Errorf("%s pll lock timeout; status 0x%x", tag, s))
			}
			time.Sleep(100 * time.Microsecond)
		}
	}
}

func (t *fe1a) Init() {
	t.CpuMain.HardwareInit()

	// Set SBUS ring map.
	{
		ringByBlock := [...]uint8{
			BlockRxPipe:    0,
			BlockLoopback0: 0, BlockLoopback1: 0, BlockLoopback2: 0, BlockLoopback3: 0,
			BlockTxPipe:  1,
			BlockMmuPipe: 2, BlockMmuSlice: 2, BlockMmuGlobal: 2,
			BlockTop: 5, BlockSer: 5, BlockAvs: 5, BlockOtpc: 5,
			BlockClport32: 3,
			BlockXlport0:  4,
		}
		// Clports 0-31
		for i := 0; i < 32; i++ {
			b := uint8(3)
			if i >= 8 && i <= 23 {
				b = 4
			}
			ringByBlock[int(BlockClport0)+i] = b
		}
		t.CpuMain.SetSbusRingMap(ringByBlock[:])
	}

	// Initialize fictitious memory-map pointers to all memories and registers.
	t.top_controller = (*top_controller)(m.BasePointer)
	t.rx_pipe_mems = (*rx_pipe_mems)(m.BasePointer)
	t.rx_pipe_controller = (*rx_pipe_controller)(m.BasePointer)
	t.tx_pipe_mems = (*tx_pipe_mems)(m.BasePointer)
	t.tx_pipe_controller = (*tx_pipe_controller)(m.BasePointer)
	t.mmu_global_mems = (*mmu_global_mems)(m.BasePointer)
	t.mmu_global_controller = (*mmu_global_controller)(m.BasePointer)
	t.mmu_pipe_mems = (*mmu_pipe_mems)(m.BasePointer)
	t.mmu_pipe_controller = (*mmu_pipe_controller)(m.BasePointer)
	t.mmu_slice_mems = (*mmu_slice_mems)(m.BasePointer)
	t.mmu_slice_controller = (*mmu_slice_controller)(m.BasePointer)

	t.dmaReq = DmaRequest{fe1a: t}
	q := t.getDmaReq()

	r := t.top_controller

	t.CoreFrequencyInHz = 850e6
	t.setCoreFreq(q)

	// Power on core pll1
	r.core_pll1_control012[0].set(q, r.core_pll1_control012[0].getDo(q)&^(1<<15))

	// Configure XG PLLs
	for i := range r.xg_pll {
		var v [2]uint32

		r.xg_pll[i].control[0].get(q, &v[0])
		r.xg_pll[i].control[4].get(q, &v[1])
		q.Do()

		// control 0: set feedback divider to divide by 20
		v[0] = (v[0] &^ (0x3ff << 18)) | (20 << 18)
		// control 0: set reference clock pre divider to 1
		v[0] = (v[0] &^ (0xf << 28)) | (1 << 28)

		// control 4: set post divider reset mode 3
		v[1] = (v[1] &^ (3 << 20)) | (3 << 20)

		r.xg_pll[i].control[0].set(q, v[0])
		r.xg_pll[i].control[4].set(q, v[1])
		q.Do()
	}

	// Configure TS PLL.
	{
		var v [1]uint32

		for i := range v {
			r.ts_pll.control[i].get(q, &v[i])
		}
		q.Do()

		// Post resetb select => post resetb
		v[0] = (v[0] &^ (3 << 24)) | (3 << 24)

		for i := range v {
			r.ts_pll.control[i].set(q, v[i])
		}
		q.Do()
	}

	// Configure BS PLL 0 & 1.
	{
		var v [2][3]uint32

		r.bs_pll0.control[0].get(q, &v[0][0])
		r.bs_pll1.control[0].get(q, &v[1][0])
		r.bs_pll0.control[2].get(q, &v[0][1])
		r.bs_pll1.control[2].get(q, &v[1][1])
		r.bs_pll0.control[3].get(q, &v[0][2])
		r.bs_pll1.control[3].get(q, &v[1][2])
		q.Do()

		// Post resetb select => post resetb
		v[0][0] = (v[0][0] &^ (3 << 24)) | (3 << 24)
		v[1][0] = (v[1][0] &^ (3 << 24)) | (3 << 24)

		// Channel 0 post-divider ratio: 125
		v[0][1] = (v[0][1] &^ (0xff << 21)) | (125 << 21)
		v[1][1] = (v[1][1] &^ (0xff << 21)) | (125 << 21)

		v[0][2] = (v[0][2] &^ (0x3ff << 0)) | (100 << 0)
		v[1][2] = (v[1][2] &^ (0x3ff << 0)) | (100 << 0)

		r.bs_pll0.control[0].set(q, v[0][0])
		r.bs_pll1.control[0].set(q, v[1][0])
		r.bs_pll0.control[2].set(q, v[0][1])
		r.bs_pll1.control[2].set(q, v[1][1])
		r.bs_pll0.control[3].set(q, v[0][2])
		r.bs_pll1.control[3].set(q, v[1][2])
		q.Do()
	}

	// Take 4 XG PLLs out of reset.
	// Also BS PLL 0-1 and TS PLL.
	soft_reset1 := r.soft_reset[1].getDo(q)
	soft_reset1 |= 1<<(2*0) | 1<<(2*1) | 1<<(2*2) | 1<<(2*3) | 1<<(2*4) | 1<<(2*5) | 1<<(2*6)
	r.soft_reset[1].set(q, soft_reset1)
	q.Do()

	// Wait for PLL lock.
	t.pll_wait(&r.xg_pll[0].status, "xg0")
	t.pll_wait(&r.xg_pll[1].status, "xg1")
	t.pll_wait(&r.xg_pll[2].status, "xg2")
	t.pll_wait(&r.xg_pll[3].status, "xg3")
	t.pll_wait(&r.ts_pll.status, "ts")
	t.pll_wait(&r.bs_pll0.status, "bs0")
	t.pll_wait(&r.bs_pll1.status, "bs1")

	// Take 4 XG + TS + BS 0-1 PLL post dividers out of reset.
	soft_reset1 |= 2<<(2*0) | 2<<(2*1) | 2<<(2*2) | 2<<(2*3) | 2<<(2*4) | 2<<(2*5) | 2<<(2*6)
	r.soft_reset[1].set(q, soft_reset1)

	// Take everything else out of reset.
	outOfReset := ^uint32(0)
	r.soft_reset[0].set(q, outOfReset)
	r.port_reset.set(q, outOfReset)
	q.Do()

	// Read revision id once and for all.
	t.revision_id = revision_id_u32(r.revision_id.getDo(q))

	t.clear_memories()
}

type revision_id_u32 uint32
type revision_id uint8

const (
	revision_id_a0 revision_id = 1
	revision_id_b0 revision_id = 0x11
	revision_id_b1 revision_id = 0x12
)

func (i revision_id_u32) getId() revision_id {
	return revision_id((i >> 16) & 0xff)
}

func (i revision_id_u32) String() string {
	return fmt.Sprintf("0x%x", i.getId())
}

func RegisterDeviceIDs(v *vnet.Vnet) {
	th := &fe1a{}
	th.SwitchConfig = m.SwitchConfig{
		PciFunction:           0,
		PhyReferenceClockInHz: 156.25e6,
	}
	m.RegisterDeviceIDs(v, th, []pci.VendorDeviceID{0xb960, 0xb961, 0xb962, 0xb963, 0xb968})
}

func (t *fe1a) String() (s string) {
	s = fmt.Sprintf("%s: id %s, rev %s",
		t.PciDev.Addr.String(), t.PciDev.DeviceID().String(),
		t.revision_id.String())
	return
}
