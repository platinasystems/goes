// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package port

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"fmt"
)

type phy_reg_dma_cmd struct {
	sbus.DmaCmd

	index uint16

	write_data uint16
	write_mask uint16

	// Storage for Rx/Tx.
	buf [4]uint32

	*phy_reg_dma_rw
}

func (p *PortBlock) phy_address(isPMD bool, lane_mask m.LaneMask, reg_offset uint16) (a uint32) {
	a = uint32(reg_offset)
	if isPMD {
		a |= 1 << 27 // DEVAD
	}
	sw := p.Switch
	phyId, _ := sw.PhyIDForPort(p.SbusBlock, 0)
	a |= uint32(phyId) << 19

	switch lane_mask {
	case 0x1 << 0:
		a |= 0 << 16
	case 0x1 << 1:
		a |= 1 << 16
	case 0x1 << 2:
		a |= 2 << 16
	case 0x1 << 3:
		a |= 3 << 16
	// Multiple lane broadcast does not reliably work for all serdes registers.
	// case 0x3 << 0:
	// 	a |= 4 << 16
	// case 0x3 << 2:
	// 	a |= 5 << 16
	// case 0xf:
	// 	a |= 6 << 16
	default:
		panic(fmt.Errorf("bad lane mask 0x%x", lane_mask))
	}

	return
}

func (c *phy_reg_dma_cmd) address() uint32 {
	return c.portBlock.phy_address(c.is_pmd_reg, c.lane_mask, c.reg_offset+c.index)
}

func (c *phy_reg_dma_cmd) Pre() {
	if c.Command.Opcode == sbus.WriteMemory {
		c.Tx = c.buf[:]

		c.Tx[0] = c.address()

		// Hardware performs: (read_value &^ write_mask) | (write_value & write_mask)
		// Ignored for reads.
		c.Tx[1] = uint32(c.write_data)<<16 | (0xffff &^ uint32(c.write_mask))

		is_write := uint32(0)
		if c.is_write {
			is_write = 1
		}
		c.Tx[2] = is_write
	} else {
		c.Rx = c.buf[:]
	}
}

func (c *phy_reg_dma_cmd) Post() {
	if !c.is_write && c.Command.Opcode == sbus.ReadMemory {
		v := uint16(c.Rx[1])
		if c.result32 != nil {
			if c.index == 0 {
				*c.result32 = uint32(v)
			} else {
				*c.result32 |= uint32(v) << 16
			}
		}
		if c.result16 != nil {
			*c.result16 = v
		}
	}
}

type phy_reg_dma_rw struct {
	portBlock *PortBlock

	is_write   bool
	is_pmd_reg bool

	reg_offset uint16

	lane_mask m.LaneMask

	result32 *uint32
	result16 *uint16
}

func (rw *phy_reg_dma_rw) get_set(q *sbus.DmaRequest, isSet bool, index, write_data, write_mask uint16) {
	cmds := [2]phy_reg_dma_cmd{}
	m := get_xclport_mems()
	cmds[0] = phy_reg_dma_cmd{
		phy_reg_dma_rw: rw,
		index:          index,
		write_data:     write_data,
		write_mask:     write_mask,
		DmaCmd: sbus.DmaCmd{
			Command: sbus.Command{Opcode: sbus.WriteMemory, Block: rw.portBlock.SbusBlock},
			Address: m.wc_ucmem_data[0].Address(),
		},
	}
	q.Add(&cmds[0])
	if !isSet {
		cmds[1] = cmds[0]
		cmds[1].Command.Opcode = sbus.ReadMemory
		q.Add(&cmds[1])
	}
}

func (p *PortBlock) GetPhyReg(q *sbus.DmaRequest, isPMD bool, lane_mask m.LaneMask, address uint16, value *uint16) {
	rw := phy_reg_dma_rw{
		portBlock:  p,
		lane_mask:  lane_mask,
		is_pmd_reg: isPMD,
		reg_offset: address,
		result16:   value,
	}
	rw.get_set(q, false, 0, 0, 0)
}

func (p *PortBlock) SetPhyReg(q *sbus.DmaRequest, isPMD bool, lane_mask m.LaneMask, address uint16, write_data, write_mask uint16) {
	rw := phy_reg_dma_rw{
		portBlock:  p,
		lane_mask:  lane_mask,
		is_pmd_reg: isPMD,
		reg_offset: address,
		is_write:   true,
	}
	rw.get_set(q, true, 0, write_data, write_mask)
}

func (p *PortBlock) GetPhyReg32(q *sbus.DmaRequest, isPMD bool, lane_mask m.LaneMask, address uint16, value *uint32) {
	rw := phy_reg_dma_rw{
		portBlock:  p,
		lane_mask:  lane_mask,
		is_pmd_reg: isPMD,
		reg_offset: address,
		result32:   value,
	}
	// Lo bits first for 32 bit registers where lo bit read triggers hardware read.
	rw.get_set(q, false, 0, 0, 0)
	rw.get_set(q, false, 1, 0, 0)
}

func (p *PortBlock) SetPhyReg32(q *sbus.DmaRequest, isPMD bool, lane_mask m.LaneMask, address uint16, value uint32) {
	rw := phy_reg_dma_rw{
		portBlock:  p,
		lane_mask:  lane_mask,
		is_pmd_reg: isPMD,
		reg_offset: address,
		is_write:   true,
	}
	// Lo bits last for 32 bit registers where lo bit write triggers hardware write.
	rw.get_set(q, true, 1, uint16(value>>16), 0xffff)
	rw.get_set(q, true, 0, uint16(value), 0xffff)
}

func (p *PortBlock) sync_rw(is_write bool, isPMD bool, lane_mask m.LaneMask, address, write_data, write_mask uint16) (read_data uint16, err error) {
	var d [4]uint32
	d[0] = p.phy_address(isPMD, lane_mask, address)
	d[1] = uint32(write_data)<<16 | (0xffff &^ uint32(write_mask))
	if is_write {
		d[2] = 1
	}
	mem := get_xclport_mems()
	sw := p.Switch.GetSwitchCommon()
	err = sw.CpuMain.PIO.Write128(p.SbusBlock, mem.wc_ucmem_data[0].Address(), d[:])
	if err == nil && !is_write {
		err = sw.CpuMain.PIO.Read128(p.SbusBlock, mem.wc_ucmem_data[0].Address(), d[:])
		read_data = uint16(d[1])
	}
	return
}

func (p *PortBlock) GetPhyRegSync(isPMD bool, lane_mask m.LaneMask, address uint16) uint16 {
	v, err := p.sync_rw(false, isPMD, lane_mask, address, 0, 0)
	if err != nil {
		panic(err)
	}
	return v
}

func (p *PortBlock) SetPhyRegSync(isPMD bool, lane_mask m.LaneMask, address, write_data, write_mask uint16) {
	_, err := p.sync_rw(true, isPMD, lane_mask, address, write_data, write_mask)
	if err != nil {
		panic(err)
	}
}

func (p *PortBlock) LoadFirmware(q *sbus.DmaRequest, ucode_bytes []byte) {
	const wordBytes = 16

	// Pad to even 16 byte block.
	nWords := uint((len(ucode_bytes)+wordBytes-1)&^(wordBytes-1)) / wordBytes

	ucode32 := make([]uint32, nWords*wordBytes/4)
	for i := range ucode_bytes {
		o := i & 0xf
		ucode32[i/4] |= uint32(ucode_bytes[i]) << uint(8*(o%4))
	}

	mems := get_xclport_mems()
	regs, _, _, _ := p.get_regs()

	// Enable and disable uc memory access before and after downloading firmware.
	block := p.SbusBlock
	enable_mem_cmd := sbus.DmaCmd{
		Command: sbus.Command{
			Opcode: sbus.WriteRegister,
			Block:  block,
		},
		Address: regs.phy_uc_data_access_mode.address(),
	}
	enable_mem_cmds := [2]sbus.DmaCmd{enable_mem_cmd, enable_mem_cmd}
	enable_mem_cmds[0].Tx = []uint32{1}
	enable_mem_cmds[1].Tx = []uint32{0}

	q.Add(&enable_mem_cmds[0])
	q.Add(&sbus.DmaCmd{
		Command: sbus.Command{
			Opcode: sbus.WriteMemory,
			Block:  block,
			Size:   wordBytes,
		},
		Address: mems.wc_ucmem_data[0].Address(),
		Tx:      ucode32,
		Count:   nWords,
	})
	q.Add(&enable_mem_cmds[1])
}
