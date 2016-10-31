// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/go/elib/hw/pci"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"

	"fmt"
	"time"
	"unsafe"
)

type top_reg uint32

func (r *top_reg) set(q *DmaRequest, v uint32)  { q.SetReg32(v, BlockTop, r.address(), sbus.Unique0) }
func (r *top_reg) get(q *DmaRequest, v *uint32) { q.GetReg32(v, BlockTop, r.address(), sbus.Unique0) }
func (r *top_reg) getDo(q *DmaRequest) (v uint32) {
	r.get(q, &v)
	q.Do()
	return
}

func (r *top_reg) address() sbus.Address { return sbus.GenReg | sbus.Address(r.offset()) }
func (r *top_reg) offset() uint          { return uint(uintptr(unsafe.Pointer(r))-m.RegsBaseAddress) << 8 }

type top_pll_regs struct {
	// [0]:
	//   channel 0: post divider [7:0], output delay enable [8],
	//   channel 5: post divider [16:9], output delay enable [17],
	//   feedback divider radio [27:18],
	//   input reference clock pre divider [31:28]
	//
	// [1] [29:8] SSC_LIMIT
	//   Spread Spectrum Clocking limit.  The nominal feedback value N is a 30-bit number,
	//   constructed from {i_ndiv_int,i_ndiv_frac}. When SSC mode is enabled, N^, the
	//   instantaneous feedback value is incremented/decremented, changing direction when it exceeds the limits.
	//   N^ < (N - {i_ssc_limit,4'b0000})
	//   N^ > (N + {i_ssc_limit,4'b0000})
	//   If N is modified while SSC mode is enabled, the stepping direction will change as the new limits are exceeded.
	//     [7] SSC_MODE Enable for spread spectrum Clocking mode; 0=normal mode; 1=spread spectrum mode enable
	//     [6] FREF_SEL R/W input reference selection: 0 = i_refclk (CMOS), 1 = pad_Fref
	//     [5:3] TESTOUT_SEL Test output clock selection:
	//   000 = no clock
	//   001 = o_fref
	//   010 = o_clock[0]
	//   011 = o_clock[1]
	//   100 = o_clock[2]
	//   101 = o_clock[3]
	//   110 = o_clock[4]
	//   111 = o_clock[5]
	//     [2]  TESTOUT_EN R/W Test enable: 0 = test output disabled, 1 = test output enabled
	//     [1:0] LDO LDO output voltage control default value 0x1
	//   00 = 1.05V
	//   01 = 1.00V
	//   10 = 0.95V
	//   11 = 0.90V
	//
	// [2] [20:17] PDIV input reference clock pre-divider control, default is 2 (decimal)
	//     [16] HOLD_CH0 Postdivider Channel0 hold: 0 = divider operational, 1 = clock hold at "0"
	//     [15:0] SSC_STEP Spread Spectrum Clocking Step Rate
	//            When SSC mode is enabled, this value determines the value that will be added to, or subtracted from N^ every refclk.
	//            N^ = N^ +/- {14'b0,i_ssc_step}
	//
	// [3] 29:10 NDIV_FRAC Fractional part of Feedback Divider Rate (N). Default is 0 (decimal).
	//   This value, divided by 2^20, is added to i_ndiv_int to determine the effective feedback divider rate N.
	//   Concatenating these busses {i_ndiv_int, i_ndiv_frac} creates a 30-bit number sometimes referred to as the Frequency Control Word (fcw),
	//   with 10 integer bits and 20 fractional bits.
	//   9:0 NDIV_INT Integer part of Feedback Divider Rate (N). Default is 100 (decimal).
	//     f(vcoclk) = f(i_refclk) * N / P., where N = Integer part, from the table below + Fractional part.
	//       P = value of i_pdiv
	//   0 => 1024, i => i
	//
	// [4]
	//   enable vco output [0], power on cml buffer 0/1 [2:1]
	//   mux test select [5:3]
	//   test buf power on [6]
	//   output level control of internal ldo [12:7]
	//   pwm rate [14:13]
	//   tdc bang bang mode [15]
	//   tdc offset control [19:16]
	//   post divider reset mode [21:20]
	control [5]top_reg

	// [31] pll is locked
	status top_reg
}

type top_regs struct {
	_ [0x2fc - 0]byte

	// port resets
	port_reset top_reg

	// [31:27] device skew
	// [26:24] chip id
	// [23:16] revision id
	// [15:0] device id (e.g. 0xb960)
	revision_id top_reg

	// All resets active low.
	// register 0: rx_pipe [0], tx pipe [1], mmu [2], ts [3], management port [4]
	// register 1: xg 0-3 ts bs 0-1 core 0-1: 9 x 2 bits: [0] reset, [1] post reset
	//             [18]/[19] temp mon/max min reset, [20] avs reset
	//   [1] defaults avs + core 0-1 out of reset (set to 1)
	soft_reset [2]top_reg

	_ [0x380 - 0x30c]byte

	core_pll0_control    [5]top_reg
	core_pll1_control012 [3]top_reg

	_ [0x400 - 0x3a0]byte

	core_pll1_control34 [2]top_reg
	core_pll_status     [2]top_reg

	_ [0x420 - 0x410]byte

	xg_pll [4]top_pll_regs

	_ [0x498 - 0x480]byte

	ts_pll                         top_pll_regs
	bs_pll0                        top_pll_regs
	l1_received_clock_valid_status [3]top_reg
	bs_pll1                        top_pll_regs

	_ [0x500 - 0x4ec]byte

	temperature_sensor struct {
		control [2]top_reg
		// current [9:0] min [21:12] max [31:22]; value v => 410.04 - v*.48705 degrees Celcius
		current [9]top_reg
	}

	_ [0x600 - 0x52c]byte

	// [31:0] tsc disable
	tsc_disable top_reg

	// [31:0] tsc afe pll lock
	tsc_afe_pll_status top_reg

	_ [0x75c - 0x608]byte

	// [25:10] reserved counter threshold
	// [9] reserved hw control enable
	// [8] sw mode mux sel
	// [7] sw core clock 1 select enable
	// [6:4] core clock 1 frequency
	//   0 => debug(950Mhz), 1 => 850Mhz, 2 => 765.625Mhz, 3 => 672Mhz, 4 => 645Mhz, 5 => 545Mhz
	// [3] sw core clock 0 select enable
	// [2:0] core clock 0 frequency
	core_pll_frequency_select top_reg

	_ [0x77c - 0x760]byte

	l1_received_clock_valid_status34 top_reg
	misc_control_2                   top_reg
	freq_switch_status               top_reg

	// ports 0-127
	port_enable            [4]top_reg
	management_port_enable top_reg
	tsc_disable_status     top_reg

	_ [0x800 - 0x7a0]byte

	temperature_sensor_interrupt struct {
		// [0] min 0 [1] max 0, [2] min 1 [3] max 1 etc.
		enable     top_reg
		thresholds [9]top_reg
		status     top_reg
	}

	_ [0x87c - 0x82c]byte

	tsc_resolved_speed_status [32]top_reg

	_ [0x914 - 0x8fc]byte

	// [7:0] interrupt status clear
	// [8:7] sram mon n process
	// [10:9] sram mon p process
	// [11] sram mon valid
	// [12] vtrap enable
	// [15:13] pvtmon bg
	// [18:16] pvtmon ref max
	// [22:19] pvtmon ref min0
	// [25:23] pvtmon ref min1
	// [26] avs disable
	// [27] cpu2avx tap enable
	avs_control top_reg
}

func (t *tomahawk) setCoreFreq(q *DmaRequest) {
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
	v := t.top_regs.core_pll_frequency_select.getDo(q)
	v |= (f << 0) | core_clock0_sw_select_enable
	t.top_regs.core_pll_frequency_select.set(q, v)
	q.Do()
}

type pipe struct {
	tdm_pipe
}

type tomahawk struct {
	m.SwitchCommon

	dmaReq DmaRequest

	revision_id revision_id_reg

	top_regs        *top_regs
	rx_pipe_mems    *rx_pipe_mems
	rx_pipe_regs    *rx_pipe_regs
	tx_pipe_mems    *tx_pipe_mems
	tx_pipe_regs    *tx_pipe_regs
	mmu_global_regs *mmu_global_regs
	mmu_global_mems *mmu_global_mems
	mmu_xpe_regs    *mmu_xpe_regs
	mmu_xpe_mems    *mmu_xpe_mems
	mmu_sc_regs     *mmu_sc_regs
	mmu_sc_mems     *mmu_sc_mems

	mmu_pipe_by_phys_port [n_phys_ports]mmu_pipe

	pipes [n_pipe]pipe

	adjacency_main
	flex_counter_main
	ip4_fib_main
	l3_main
	l3_interface_main
	my_station_tcam_main
	portMain
	port_counter_main
	port_bitmap_main
	source_trunk_map_main
}

func (t *tomahawk) getDmaReq() *DmaRequest { return &t.dmaReq }
func (t *tomahawk) Interrupt()             { t.Cmic.Interrupt() }
func (t *tomahawk) GetLedDataRam()         { t.Cmic.LedDataRamDump() }

func (t *tomahawk) pll_wait(r *top_reg, tag string) {
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

func (t *tomahawk) Init() {
	t.Cmic.HardwareInit()

	// Set SBUS ring map.
	{
		ringByBlock := [...]uint8{
			BlockRxPipe:    0,
			BlockLoopback0: 0, BlockLoopback1: 0, BlockLoopback2: 0, BlockLoopback3: 0,
			BlockTxPipe: 1,
			BlockMmuXpe: 2, BlockMmuSc: 2, BlockMmuGlobal: 2,
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
		t.Cmic.SetSbusRingMap(ringByBlock[:])
	}

	// Initialize fictitious memory-map pointers to all memories and registers.
	t.top_regs = (*top_regs)(m.RegsBasePointer)
	t.rx_pipe_mems = (*rx_pipe_mems)(m.RegsBasePointer)
	t.rx_pipe_regs = (*rx_pipe_regs)(m.RegsBasePointer)
	t.tx_pipe_mems = (*tx_pipe_mems)(m.RegsBasePointer)
	t.tx_pipe_regs = (*tx_pipe_regs)(m.RegsBasePointer)
	t.mmu_global_mems = (*mmu_global_mems)(m.RegsBasePointer)
	t.mmu_global_regs = (*mmu_global_regs)(m.RegsBasePointer)
	t.mmu_xpe_mems = (*mmu_xpe_mems)(m.RegsBasePointer)
	t.mmu_xpe_regs = (*mmu_xpe_regs)(m.RegsBasePointer)
	t.mmu_sc_mems = (*mmu_sc_mems)(m.RegsBasePointer)
	t.mmu_sc_regs = (*mmu_sc_regs)(m.RegsBasePointer)

	t.dmaReq = DmaRequest{tomahawk: t}
	q := t.getDmaReq()

	r := t.top_regs

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
	t.revision_id = revision_id_reg(r.revision_id.getDo(q))

	t.clear_memories()
}

type revision_id_reg uint32
type revision_id uint8

const (
	revision_id_a0 revision_id = 1
	revision_id_b0 revision_id = 0x11
	revision_id_b1 revision_id = 0x12
)

func (i revision_id_reg) getId() revision_id {
	return revision_id((i >> 16) & 0xff)
}

func (i revision_id_reg) String() string {
	return fmt.Sprintf("0x%x", i.getId())
}

func Init(v *vnet.Vnet) {
	th := &tomahawk{}
	th.SwitchConfig = m.SwitchConfig{
		PciFunction:           0,
		PhyReferenceClockInHz: 156.25e6,
	}
	m.RegisterDeviceIDs(v, th, []pci.VendorDeviceID{0xb960, 0xb961, 0xb962, 0xb963, 0xb968})
}

func (t *tomahawk) String() (s string) {
	s = fmt.Sprintf("%s: id %s, rev %s",
		t.PciDev.Addr.String(), t.PciDev.DeviceID().String(),
		t.revision_id.String())
	return
}
