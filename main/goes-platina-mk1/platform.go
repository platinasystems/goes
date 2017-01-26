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
