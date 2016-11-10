// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpu

import (
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/i2c"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/iproc"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/led"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/mdio"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/packet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"fmt"
	"time"
	"unsafe"
)

type cmc_regs struct {
	schan      sbus.SchanRegs
	fast_schan sbus.FastSchanRegs

	_ [0x110 - 0x80]byte

	packet_dma packet.DmaRegs

	_ [0x2c0 - 0x1c0]byte

	fifo_dma sbus.FifoDmaRegs

	_ [0x400 - 0x3a0]byte

	irq_status0       [5]hw.Reg32
	irq_enable0       [3][5]hw.Reg32
	irq_enable_rcpu   hw.Reg32
	ecc_errors_status [2]hw.Reg32

	_ [0x470 - 0x45c]byte

	pcie_msi_config hw.Reg32

	hostmem_addr_remap0 [3]hw.Reg32

	packet_counts [5]struct{ rx, tx hw.Reg32 }

	sw_irq_config hw.Reg32

	hostmem_addr_remap1 [1]hw.Reg32
	irq_status1         [2]hw.Reg32
	irq_enable1         [3][2]hw.Reg32

	_ [0x600 - 0x4d0]byte

	sbus_dma                 [3]sbus.DmaRegs
	sbus_dma_timers          [3]hw.Reg32
	subs_dma_iteration_count [3]hw.Reg32

	_ [0x1000 - 0x708]byte
}

type regs struct {
	i2c iproc.I2cRegs

	_ [0x10c - 0x50]byte

	config hw.Reg32

	_ [0x10000 - 0x110]byte

	schan sbus.SchanRegs

	_ [0x10080 - 0x10064]byte

	mdio mdio.Regs

	sbus struct {
		timeout hw.Reg32

		ring_map [16]hw.Reg32
	}
	_             [0x101ec - 0x100d8]byte
	endian_config struct {
		pcie hw.Reg32
		uc   [2]hw.Reg32
		spi  hw.Reg32
		i2c  hw.Reg32
		rpe  hw.Reg32
	}
	pio_mcs_access_page hw.Reg32

	pcie_config hw.Reg32

	uc_config               [2]hw.Reg32
	pcie_error_status       hw.Reg32
	pcie_error_status_clear hw.Reg32
	sw_reset                hw.Reg32

	cps_reset hw.Reg32

	revision_id hw.Reg32

	cpu_interface_revision_id hw.Reg32
	_                         hw.Reg32
	pcie_reset_control        hw.Reg32

	override_strap hw.Reg32

	bspi_bigendian hw.Reg32

	_ hw.Reg32

	strap_status         [2]hw.Reg32
	fsrf_standby_control hw.Reg32

	_ hw.Reg32

	pcie_user_if_timeout      hw.Reg32
	pcie_user_if_status       hw.Reg32
	pcie_user_if_status_clear hw.Reg32
	pcie_user_if_status_mask  hw.Reg32

	pcie_user_if_purge_control hw.Reg32

	pcie_user_if_purge_status hw.Reg32
	pcie_dma_buf_mem_control  hw.Reg32
	mcs_dma_buf_mem_control   hw.Reg32
	pcie_address_upper_32bits hw.Reg32

	_ [0x11000 - 0x10274]byte

	miim miim_regs

	_ [0x1a000 - 0x1123c]byte

	rx_buf struct {
		epipe_interface_release_all_credits hw.Reg32

		epipe_interface_max_interface_credits hw.Reg32

		status_buffer_max_free_list_entries hw.Reg32

		data_buffer_max_free_list_entries hw.Reg32

		status_buffer_n_free_entries hw.Reg32
		data_buffer_n_free_entries   hw.Reg32

		status_buffer_alloc hw.Reg32
		data_buffer_alloc   hw.Reg32

		epipe_interface_buffer_depth hw.Reg32

		config            hw.Reg32
		ecc_error_control hw.Reg32
		data_buffer_tm    [3]hw.Reg32
		status_buffer_tm  [2]hw.Reg32

		_ [0x1b000 - 0x1a040]byte
	}

	tx_buf struct {
		max_buffer_limits hw.Reg32
		min_buffer_limits hw.Reg32

		_ hw.Reg32

		packet_counts [4]hw.Reg32 // rpe, cmc0-2
		debug         hw.Reg32

		_ hw.Reg32

		ipipe_interface_buffer_depth hw.Reg32

		data_buffer_n_free_entries hw.Reg32

		ipipe_interface_credits hw.Reg32

		data_buffer_max_free_list_entries hw.Reg32

		config hw.Reg32

		status       hw.Reg32
		status_clear hw.Reg32

		ecc_error_control hw.Reg32
		data_buffer_tm    [2]hw.Reg32
		mhdr_buffer_tm    [1]hw.Reg32
	}

	_ [0x20000 - 0x1b050]byte

	led0 led.Regs
	led1 led.Regs
	_    [0x29000 - 0x22000]byte
	led2 led.Regs
	_    [0x31000 - 0x2a000]byte
	cmc  [3]cmc_regs
}

type Config struct {
	SbusTimeout     uint32
	SbusRingByBlock map[int]int
}

type Main struct {
	regs      *regs
	iprocRegs *iproc.Regs
	Config    Config

	sbus.Sbus
	mdio.Mdio
	PacketDma packet.Dma

	i2c.I2c

	Leds [3]led.Led

	// Interrupt handlers indexed by interrupt number.
	interruptHandlers interruptHandlerVec
	interruptCount    uint64

	linkScanMain
}

func (cm *Main) SetSbusRingMap(data []uint8) {
	var d [128 / 8]uint32
	for i := range data {
		d[i/8] |= (uint32(data[i]) & 0xf) << (4 * (uint(i) % 8))
	}
	for i := range d {
		cm.regs.sbus.ring_map[i].Set(d[i])
	}
}

func (cm *Main) Reset() {
	r := cm.regs

	// Hard reset of chip.
	start := time.Now()
	r.cps_reset.Set(1)
	for r.cps_reset.Get()&1 != 0 {
		if time.Since(start) > 100*time.Millisecond {
			panic("cps reset timeout")
		}
	}

	cf := &cm.Config

	if cf.SbusTimeout == 0 {
		cf.SbusTimeout = 0x7d0
	}
	r.sbus.timeout.Set(cf.SbusTimeout)

	if cm.iprocRegs != nil {
		a := cm.iprocRegs.Init()
		addr33_28 := uint32(a >> 28)
		var v [4]uint32
		for i := uint32(0); i < 16; i++ {
			v[i/5] |= (addr33_28 | i) << uint(6*(i%5))
		}
		x := &r.cmc[0]
		x.hostmem_addr_remap0[0].Set(v[0])
		x.hostmem_addr_remap0[1].Set(v[1])
		x.hostmem_addr_remap0[2].Set(v[2])
		x.hostmem_addr_remap1[0].Set(v[3])
	}

	cm.interrupt_init()
}

func (cm *Main) Init(pRegs, pIproc unsafe.Pointer) {
	cm.regs = (*regs)(pRegs)
	cm.iprocRegs = (*iproc.Regs)(pIproc)

	c := &cm.regs.cmc[0]

	cm.Schan.Regs = &c.schan
	cm.Schan.FastRegs = &c.fast_schan
	cm.setInterruptHandler(schan_done_interrupt, cm.Schan.DoneInterrupt)

	cm.Dma.InitChannels(c.sbus_dma[:])
	cm.setInterruptHandler(sbus_dma0_interrupt, cm.Dma.Channels[0].Interrupt)
	cm.setInterruptHandler(sbus_dma1_interrupt, cm.Dma.Channels[1].Interrupt)
	cm.setInterruptHandler(sbus_dma2_interrupt, cm.Dma.Channels[2].Interrupt)

	cm.PacketDma.InitChannels(&c.packet_dma)
	cm.PacketDma.InterruptEnable = cm.packetDmaInterruptEnable
	cm.setInterruptHandler(packet_dma0_desc_controlled_interrupt, cm.PacketDma.Channels[0].DescControlledInterrupt)
	cm.setInterruptHandler(packet_dma1_desc_controlled_interrupt, cm.PacketDma.Channels[1].DescControlledInterrupt)
	cm.setInterruptHandler(packet_dma2_desc_controlled_interrupt, cm.PacketDma.Channels[2].DescControlledInterrupt)
	cm.setInterruptHandler(packet_dma3_desc_controlled_interrupt, cm.PacketDma.Channels[3].DescControlledInterrupt)

	cm.FifoDma.InitChannels(&c.fifo_dma)
	cm.setInterruptHandler(fifo_dma0_interrupt, cm.FifoDma.Channels[0].Interrupt)
	cm.setInterruptHandler(fifo_dma1_interrupt, cm.FifoDma.Channels[1].Interrupt)
	cm.setInterruptHandler(fifo_dma2_interrupt, cm.FifoDma.Channels[2].Interrupt)
	cm.setInterruptHandler(fifo_dma3_interrupt, cm.FifoDma.Channels[3].Interrupt)

	cm.Mdio.Regs = &cm.regs.mdio
	cm.setInterruptHandler(mdio_done_interrupt, cm.Mdio.DoneInterrupt)

	cm.setInterruptHandler(phy_scan_link_status_interrupt, cm.LinkStatusChangeInterrupt)
	cm.setInterruptHandler(phy_scan_pause_status_interrupt, cm.PauseStatusChangeInterrupt)

	cm.I2cRegs = &cm.regs.i2c
	cm.setInterruptHandler(i2c_interrupt, cm.I2c.Interrupt)
}

func (cm *Main) HardwareInit() {
	cm.Reset()

	// Override master mode on i2c.
	cm.regs.override_strap.Set(cm.regs.override_strap.Get() | (1 << 1) | (1 << 4))
	cm.I2c.Init()
}

func (cm *Main) packetDmaInterruptEnable(enable bool) {
	const mask = 1<<packet_dma0_desc_controlled_interrupt |
		1<<packet_dma1_desc_controlled_interrupt |
		1<<packet_dma2_desc_controlled_interrupt |
		1<<packet_dma3_desc_controlled_interrupt
	cm.interruptEnable(0, mask, enable)
}

func (cm *Main) StartPacketDma() {
	cm.regs.rx_buf.epipe_interface_release_all_credits.Set(1)
}

func (cm *Main) LedInit(data_ram_port_offset uint) {
	// FIXME skip this when there are no leds.
	cm.Leds[0].Regs = &cm.regs.led0
	cm.Leds[1].Regs = &cm.regs.led1
	cm.Leds[2].Regs = &cm.regs.led2
	for i := range cm.Leds {
		cm.Leds[i].Init(data_ram_port_offset, i)
	}
}

func (cm *Main) LedDataRamDump() {
	cm.Leds[0].Regs = &cm.regs.led0
	cm.Leds[1].Regs = &cm.regs.led1
	for i := uint32(0); i < 2; i++ {
		fmt.Printf("LED Data Ram[%d]\n", i)
		cm.Leds[i].DumpDataRam()
	}
}
