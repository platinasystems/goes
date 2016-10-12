package ixge

import (
	"github.com/platinasystems/go/vnet/devices/phy/xge"
)

func (d *dev) rw_phy_reg(dev_type, reg_index, v reg, is_read bool) (w reg) {
	const busy_bit = 1 << 30
	sync_mask := reg(1) << (1 + d.phy_index)
	d.software_firmware_sync(sync_mask, 0)
	if !is_read {
		d.regs.xge_mac.phy_data.set(d, v)
	}
	// Address cycle.
	x := reg_index | dev_type<<16 | d.phys[d.phy_index].mdio_address<<21
	d.regs.xge_mac.phy_command.set(d, x|busy_bit)
	for d.regs.xge_mac.phy_command.get(d)&busy_bit != 0 {
	}
	cmd := reg(1)
	if is_read {
		cmd = 2
	}
	d.regs.xge_mac.phy_command.set(d, x|busy_bit|cmd<<26)
	for d.regs.xge_mac.phy_command.get(d)&busy_bit != 0 {
	}
	if is_read {
		w = d.regs.xge_mac.phy_data.get(d) >> 16
	} else {
		w = v
	}
	d.software_firmware_sync_release(sync_mask, 0)
	return
}

func (d *dev) read_phy_reg(dev_type, reg_index reg) reg {
	return d.rw_phy_reg(dev_type, reg_index, 0, true)
}
func (d *dev) write_phy_reg(dev_type, reg_index, v reg) {
	d.rw_phy_reg(dev_type, reg_index, v, false)
}

func (d *dev) probe_phy() (ok bool) {
	phy := &d.phys[d.phy_index]

	phy.mdio_address = ^phy.mdio_address // poison
	for i := reg(0); i < 32; i++ {
		phy.mdio_address = i
		v := d.read_phy_reg(xge.PHY_DEV_TYPE_PMA_PMD, xge.PHY_ID1)
		if ok = v != 0xffff && v != 0; ok {
			phy.id = uint32(v)
			break
		}
	}
	return
}

func (d *dev) rw_iphy_reg(dev_type, reg_index, v reg, is_read bool) (w reg) {
	sync_mask := reg(1) << (1 + d.phy_index)
	d.software_firmware_sync(sync_mask, 0)
	d.regs.indirect_phy.control.set(d, reg_index|dev_type<<28)
	if !is_read {
		d.regs.indirect_phy.data.set(d, v)
		w = v
	} else {
		w = d.regs.indirect_phy.data.get(d)
	}
	const busy_bit = 1 << 31
	for d.regs.indirect_phy.control.get(d)&busy_bit != 0 {
	}
	d.software_firmware_sync_release(sync_mask, 0)
	return
}

func (d *dev) read_iphy_reg(dev_type, reg_index reg) reg {
	return d.rw_iphy_reg(dev_type, reg_index, 0, true)
}
func (d *dev) write_iphy_reg(dev_type, reg_index, v reg) {
	d.rw_iphy_reg(dev_type, reg_index, v, false)
}
