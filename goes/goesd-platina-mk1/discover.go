// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/eeprom"
	"github.com/platinasystems/go/i2c"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm"
)

func (p *platform) Init() (err error) {
	v := p.Vnet
	p.Platform = bcm.GetPlatform(v)
	if err = p.boardInit(); err != nil {
		v.Logf("boardInit failure: %s", err)
		return
	}
	for _, s := range p.Switches {
		if err = p.boardPortInit(s); err != nil {
			v.Logf("boardPortInit failure: %s", err)
			return
		}
	}

	if err = p.boardPortLedEnable(); err != nil {
		v.Logf("boardPortLedEnable failure: %s", err)
	}

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

func (p *platform) boardInit() (err error) {
	// The MK1 x86 CPU Card EEPROM is located on bus 0, addr 0x51:
	// Read and store the EEPROM Contents
	d := eeprom.Device{
		BusIndex:   0,
		BusAddress: 0x51,
	}
	if e := d.GetInfo(); e != nil {
		p.Vnet.Logf("eeprom read failed: %s; using random addresses", err)
		p.Platform.AddressBlock = ethernet.AddressBlock{
			Base:  ethernet.RandomAddress(),
			Count: 256,
		}
	} else {
		// in case the eeprom read fails
		p.Platform.AddressBlock = ethernet.AddressBlock{
			Base:  d.Fields.BaseEthernetAddress,
			Count: uint32(d.Fields.NEthernetAddress),
		}
	}
	return
}

func (p *platform) boardPortInit(s bcm.Switch) (err error) {
	cf := bcm.SwitchConfig{}

	const (
		numPortsDefault = 32
	)

	cf.Ports = make([]bcm.PortConfig, numPortsDefault)

	// Data ports
	for i := range cf.Ports {
		cf.Ports[i] = bcm.PortConfig{
			PortBlockIndex:  uint(i),
			SpeedBitsPerSec: 100e9,
			LaneMask:        0xf,
			PhyInterface:    bcm.PhyInterfaceOptics,
		}
	}

	// Management ports.
	for i := uint(0); i < 2; i++ {
		cf.Ports = append(cf.Ports, bcm.PortConfig{
			PortBlockIndex:  0,
			SubPortIndex:    i,
			IsManagement:    true,
			SpeedBitsPerSec: 10e9,
			LaneMask:        1 << (2 * uint(i)),
			PhyInterface:    bcm.PhyInterfaceKR,
		})
	}

	phys := [32 + 1]bcm.PhyConfig{}
	// No front panel remapping; no lane remapping on Mk1
	for i := range phys {
		p := &phys[i]
		p.Index = uint8(i & 0x1f)
		p.FrontPanelIndex = p.Index
		p.IsManagement = i == 32
	}
	cf.Phys = phys[:]

	cf.Configure(p.Vnet, s)
	return
}
