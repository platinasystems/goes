// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/log"
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
		v.Logf("boardInit failure: %s", err)
		return
	}
	for _, s := range p.Switches {
		if err = p.boardPortInit(s); err != nil {
			v.Logf("boardPortInit failure: %s", err)
			return
		}
	}

	if len(p.Switches) > 0 {
		// don't need led enable if we're not running on hardware.
		if err = p.boardPortLedEnable(); err != nil {
			v.Logf("boardPortLedEnable failure: %s", err)
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

func (p *platform) boardInit() (err error) {
	var s string
	var macs uint
	var ethAddr [6]byte
	defer func() {
		log.Printf("MK1 x86 eeprom MAC addresses: MAC_BASE: [% x], #MAC's: 0x%x (%d)\n", ethAddr, macs, macs)
		if err == nil {
			p.Platform.AddressBlock = ethernet.AddressBlock{
				Base:  ethAddr,
				Count: uint32(macs),
			}
		} else { // eeprom values are invalid or not programmed correctly
			log.Printf("Exiting: Invalid or incorrectly programmed MK1 x86 eeprom\n")
			panic(err)
		}
	}()

	if s, err = redis.Hget(redis.DefaultHash, "eeprom.number_of_ethernet_addrs"); err != nil {
		return
	}
	if _, err = fmt.Sscan(s, &macs); err != nil && macs < 134 { // at least 134 MAC addresses..
		return
	}
	if s, err = redis.Hget(redis.DefaultHash, "eeprom.base_ethernet_address"); err != nil {
		return
	}
	str := strings.Fields(s[1:(len(s) - 1)]) // remove the '[ ]' brackets from the string
	for i := range str {
		var u8 uint8
		if _, err = fmt.Sscan(str[i], &u8); err != nil {
			return
		}
		ethAddr[i] = u8
	}
	return
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

	for i := range phys {
		p := &phys[i]
		p.Index = uint8(i & 0x1f)
		if devEeprom.Fields.DeviceVersion == 0 {
			// Alpha level board
			// No lane remapping, but the MK1 front panel ports are flipped and 0-based.
			p.FrontPanelIndex = p.Index ^ 1
		} else {
			// Beta & Production level boards have version 1 and above
			// No lane remapping, but the MK1 front panel ports are flipped and 1-based.
			p.FrontPanelIndex = (p.Index ^ 1) + 1
		}
		p.IsManagement = i == 32
	}
	cf.Phys = phys[:]

	cf.Configure(p.Vnet, s)
	return
}
