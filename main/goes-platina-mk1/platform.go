// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"

	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/optional/vnetd"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1"
	"github.com/platinasystems/go/vnet/devices/optics/qsfp"
	"github.com/platinasystems/go/vnet/ethernet"
)

type platform struct {
	vnet.Package
	*fe1.Platform
	i *vnetd.Info
}

func (p *platform) Init() (err error) {
	v := p.Vnet
	p.Platform = fe1.GetPlatform(v)
	if err = p.boardInit(); err != nil {
		v.Logf("boardInit failure: %s\n", err)
		return
	}
	for _, s := range p.Switches {
		if err = p.boardPortInit(s); err != nil {
			v.Logf("boardPortInit failure: %s\n", err)
			return
		}
	}

	if len(p.Switches) > 0 {
		// don't need led enable if we're not running on hardware.
		if err = p.boardPortLedEnable(); err != nil {
			v.Logf("boardPortLedEnable failure: %s\n", err)
		}
	}

	vnetd.Init(p.i)
	return
}

// MK1 board front panel port LED's require PCA9535 GPIO device
// configuration - to provide an output signal that allows LED
// operation.
func (p *platform) boardPortLedEnable() (err error) {
	var bus i2c.Bus
	var busIndex, busAddress int = 0, 0x74

	err = bus.Open(busIndex)
	if err != nil {
		return err
	}
	defer bus.Close()

	err = bus.ForceSlaveAddress(busAddress)
	if err != nil {
		return err
	}

	// Configure the gpio pin as an output:
	// Register 6 controls the configuration, bit 2 is led enable, '0' => 'output'
	const (
		pca9535ConfigReg = 0x6
		ledOutputEnable  = 1 << 2
	)
	var data i2c.SMBusData
	data[0] = ^uint8(ledOutputEnable)
	err = bus.Do(i2c.Write, uint8(pca9535ConfigReg), i2c.ByteData, &data)
	return err
}

func (p *platform) boardInit() error {
	macs, err := nMacs()
	if err != nil {
		return err
	}
	base, err := baseEtherAddr()
	if err != nil {
		return err
	}
	p.Platform.AddressBlock = ethernet.AddressBlock{
		Base:  base,
		Count: uint32(macs),
	}
	if false {
		fmt.Println("eeprom.NEthernetAddress:", macs)
		fmt.Println("eeprom.BaseEthernetAddress:", base.String())
	}
	return nil
}

func (p *platform) boardPortInit(s fe1.Switch) (err error) {
	cf := fe1.SwitchConfig{
		Ports: make([]fe1.PortConfig, 32),
	}

	// Data ports
	for i := range cf.Ports {
		cf.Ports[i] = fe1.PortConfig{
			PortBlockIndex:  uint(i),
			SpeedBitsPerSec: 100e9,
			LaneMask:        0xf,
			PhyInterface:    fe1.PhyInterfaceOptics,
		}
	}

	// Management ports.
	for i := uint(0); i < 2; i++ {
		cf.Ports = append(cf.Ports, fe1.PortConfig{
			PortBlockIndex:  0,
			SubPortIndex:    i,
			IsManagement:    true,
			SpeedBitsPerSec: 10e9,
			LaneMask:        1 << (2 * uint(i)),
			PhyInterface:    fe1.PhyInterfaceKR,
		})
	}

	phys := [32 + 1]fe1.PhyConfig{}

	ver, err := deviceVersion()
	if err != nil {
		return
	}

	// Alpha level board (version 0):
	//   No lane remapping, but the MK1 front panel ports are flipped and 0-based.
	// Beta & Production level boards have version 1 and above:
	//   No lane remapping, but the MK1 front panel ports are flipped and 1-based.
	if ver > 0 {
		p.PortNumberOffset = 1
	}

	for i := range phys {
		phy := &phys[i]
		phy.Index = uint8(i & 0x1f)
		phy.FrontPanelIndex = phy.Index ^ 1
		phy.IsManagement = i == 32
	}
	cf.Phys = phys[:]

	cf.Configure(p.Vnet, s)
	return
}

func i2cAddrs() {
	qsfpioInit()
	qsfpInit()
}

func qsfpioInit() {
	qsfp.VdevIo[0] = qsfp.I2cDev{0, 0x20, 0, 0x70, 0x10, 0, 0, 0} //port 1-16 present signals
	qsfp.VdevIo[1] = qsfp.I2cDev{0, 0x21, 0, 0x70, 0x10, 0, 0, 0} //port 17-32 present signals
	qsfp.VdevIo[2] = qsfp.I2cDev{0, 0x22, 0, 0x70, 0x10, 0, 0, 0} //port 1-16 interrupt signals
	qsfp.VdevIo[3] = qsfp.I2cDev{0, 0x23, 0, 0x70, 0x10, 0, 0, 0} //port 17-32 interrupt signals
	qsfp.VdevIo[4] = qsfp.I2cDev{0, 0x20, 0, 0x70, 0x20, 0, 0, 0} //port 1-16 LP mode signals
	qsfp.VdevIo[5] = qsfp.I2cDev{0, 0x21, 0, 0x70, 0x20, 0, 0, 0} //port 17-32 LP mode signals
	qsfp.VdevIo[6] = qsfp.I2cDev{0, 0x22, 0, 0x70, 0x20, 0, 0, 0} //port 1-16 reset signals
	qsfp.VdevIo[7] = qsfp.I2cDev{0, 0x23, 0, 0x70, 0x20, 0, 0, 0} //port 17-32 reset signals

	qsfp.VpageByKeyIo = map[string]uint8{
		"port-1.qsfp.presence":  0,
		"port-2.qsfp.presence":  0,
		"port-3.qsfp.presence":  0,
		"port-4.qsfp.presence":  0,
		"port-5.qsfp.presence":  0,
		"port-6.qsfp.presence":  0,
		"port-7.qsfp.presence":  0,
		"port-8.qsfp.presence":  0,
		"port-9.qsfp.presence":  0,
		"port-10.qsfp.presence": 0,
		"port-11.qsfp.presence": 0,
		"port-12.qsfp.presence": 0,
		"port-13.qsfp.presence": 0,
		"port-14.qsfp.presence": 0,
		"port-15.qsfp.presence": 0,
		"port-16.qsfp.presence": 0,
		"port-17.qsfp.presence": 1,
		"port-18.qsfp.presence": 1,
		"port-19.qsfp.presence": 1,
		"port-20.qsfp.presence": 1,
		"port-21.qsfp.presence": 1,
		"port-22.qsfp.presence": 1,
		"port-23.qsfp.presence": 1,
		"port-24.qsfp.presence": 1,
		"port-25.qsfp.presence": 1,
		"port-26.qsfp.presence": 1,
		"port-27.qsfp.presence": 1,
		"port-28.qsfp.presence": 1,
		"port-29.qsfp.presence": 1,
		"port-30.qsfp.presence": 1,
		"port-31.qsfp.presence": 1,
		"port-32.qsfp.presence": 1,
	}
}

func qsfpInit() {
	qsfp.Vdev[0] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x1}
	qsfp.Vdev[1] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x2}
	qsfp.Vdev[2] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x4}
	qsfp.Vdev[3] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x8}
	qsfp.Vdev[4] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x10}
	qsfp.Vdev[5] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x20}
	qsfp.Vdev[6] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x40}
	qsfp.Vdev[7] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x80}
	qsfp.Vdev[8] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x1}
	qsfp.Vdev[9] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x2}
	qsfp.Vdev[10] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x4}
	qsfp.Vdev[11] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x8}
	qsfp.Vdev[12] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x10}
	qsfp.Vdev[13] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x20}
	qsfp.Vdev[14] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x40}
	qsfp.Vdev[15] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x80}
	qsfp.Vdev[16] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x1}
	qsfp.Vdev[17] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x2}
	qsfp.Vdev[18] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x4}
	qsfp.Vdev[19] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x8}
	qsfp.Vdev[20] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x10}
	qsfp.Vdev[21] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x20}
	qsfp.Vdev[22] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x40}
	qsfp.Vdev[23] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x80}
	qsfp.Vdev[24] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x1}
	qsfp.Vdev[25] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x2}
	qsfp.Vdev[26] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x4}
	qsfp.Vdev[27] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x8}
	qsfp.Vdev[28] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x10}
	qsfp.Vdev[29] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x20}
	qsfp.Vdev[30] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x40}
	qsfp.Vdev[31] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x80}
}

func deviceVersion() (ver int, err error) {
	s, err := redis.Hget(redis.DefaultHash, "eeprom.DeviceVersion")
	if err != nil {
		return
	}
	_, err = fmt.Sscan(s, &ver)
	return
}

func nMacs() (n uint, err error) {
	s, err := redis.Hget(redis.DefaultHash, "eeprom.NEthernetAddress")
	if err != nil {
		return
	}
	_, err = fmt.Sscan(s, &n)
	return
}

func baseEtherAddr() (ea ethernet.Address, err error) {
	s, err := redis.Hget(redis.DefaultHash, "eeprom.BaseEthernetAddress")
	if err != nil {
		return
	}
	input := new(parse.Input)
	input.SetString(s)
	ea.Parse(input)
	return
}
