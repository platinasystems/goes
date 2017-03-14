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
	"github.com/platinasystems/go/vnet/devices/optics/qsfpio"
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

func (p *platform) QsfpioInit() {
	//port 1-16 present signals
	qsfpio.Vdev[0].Bus = 0
	qsfpio.Vdev[0].Addr = 0x20
	qsfpio.Vdev[0].MuxBus = 0
	qsfpio.Vdev[0].MuxAddr = 0x70
	qsfpio.Vdev[0].MuxValue = 0x10

	//port 17-32 present signals
	qsfpio.Vdev[1].Bus = 0
	qsfpio.Vdev[1].Addr = 0x21
	qsfpio.Vdev[1].MuxBus = 0
	qsfpio.Vdev[1].MuxAddr = 0x70
	qsfpio.Vdev[1].MuxValue = 0x10

	//port 1-16 interrupt signals
	qsfpio.Vdev[2].Bus = 0
	qsfpio.Vdev[2].Addr = 0x22
	qsfpio.Vdev[2].MuxBus = 0
	qsfpio.Vdev[2].MuxAddr = 0x70
	qsfpio.Vdev[2].MuxValue = 0x10

	//port 17-32 interrupt signals
	qsfpio.Vdev[3].Bus = 0
	qsfpio.Vdev[3].Addr = 0x23
	qsfpio.Vdev[3].MuxBus = 0
	qsfpio.Vdev[3].MuxAddr = 0x70
	qsfpio.Vdev[3].MuxValue = 0x10

	//port 1-16 LP mode signals
	qsfpio.Vdev[4].Bus = 0
	qsfpio.Vdev[4].Addr = 0x20
	qsfpio.Vdev[4].MuxBus = 0
	qsfpio.Vdev[4].MuxAddr = 0x70
	qsfpio.Vdev[4].MuxValue = 0x20

	//port 17-32 LP mode signals
	qsfpio.Vdev[5].Bus = 0
	qsfpio.Vdev[5].Addr = 0x21
	qsfpio.Vdev[5].MuxBus = 0
	qsfpio.Vdev[5].MuxAddr = 0x70
	qsfpio.Vdev[5].MuxValue = 0x20

	//port 1-16 reset signals
	qsfpio.Vdev[6].Bus = 0
	qsfpio.Vdev[6].Addr = 0x22
	qsfpio.Vdev[6].MuxBus = 0
	qsfpio.Vdev[6].MuxAddr = 0x70
	qsfpio.Vdev[6].MuxValue = 0x20

	//port 1-32 reset signals
	qsfpio.Vdev[7].Bus = 0
	qsfpio.Vdev[7].Addr = 0x23
	qsfpio.Vdev[7].MuxBus = 0
	qsfpio.Vdev[7].MuxAddr = 0x70
	qsfpio.Vdev[7].MuxValue = 0x20

	qsfpio.VpageByKey = map[string]uint8{
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

	return
}

func (p *platform) QsfpInit() {
	qsfp.Vdev[0].Bus = 0
	qsfp.Vdev[0].Addr = 0x50
	qsfp.Vdev[0].MuxBus = 0
	qsfp.Vdev[0].MuxAddr = 0x70
	qsfp.Vdev[0].MuxValue = 0x1
	qsfp.Vdev[0].MuxBus2 = 0
	qsfp.Vdev[0].MuxAddr2 = 0x71
	qsfp.Vdev[0].MuxValue2 = 0x1

	qsfp.Vdev[1].Bus = 0
	qsfp.Vdev[1].Addr = 0x50
	qsfp.Vdev[1].MuxBus = 0
	qsfp.Vdev[1].MuxAddr = 0x70
	qsfp.Vdev[1].MuxValue = 0x1
	qsfp.Vdev[1].MuxBus2 = 0
	qsfp.Vdev[1].MuxAddr2 = 0x71
	qsfp.Vdev[1].MuxValue2 = 0x2

	qsfp.Vdev[2].Bus = 0
	qsfp.Vdev[2].Addr = 0x50
	qsfp.Vdev[2].MuxBus = 0
	qsfp.Vdev[2].MuxAddr = 0x70
	qsfp.Vdev[2].MuxValue = 0x1
	qsfp.Vdev[2].MuxBus2 = 0
	qsfp.Vdev[2].MuxAddr2 = 0x71
	qsfp.Vdev[2].MuxValue2 = 0x4

	qsfp.Vdev[3].Bus = 0
	qsfp.Vdev[3].Addr = 0x50
	qsfp.Vdev[3].MuxBus = 0
	qsfp.Vdev[3].MuxAddr = 0x70
	qsfp.Vdev[3].MuxValue = 0x1
	qsfp.Vdev[3].MuxBus2 = 0
	qsfp.Vdev[3].MuxAddr2 = 0x71
	qsfp.Vdev[3].MuxValue2 = 0x8

	qsfp.Vdev[4].Bus = 0
	qsfp.Vdev[4].Addr = 0x50
	qsfp.Vdev[4].MuxBus = 0
	qsfp.Vdev[4].MuxAddr = 0x70
	qsfp.Vdev[4].MuxValue = 0x1
	qsfp.Vdev[4].MuxBus2 = 0
	qsfp.Vdev[4].MuxAddr2 = 0x71
	qsfp.Vdev[4].MuxValue2 = 0x10

	qsfp.Vdev[5].Bus = 0
	qsfp.Vdev[5].Addr = 0x50
	qsfp.Vdev[5].MuxBus = 0
	qsfp.Vdev[5].MuxAddr = 0x70
	qsfp.Vdev[5].MuxValue = 0x1
	qsfp.Vdev[5].MuxBus2 = 0
	qsfp.Vdev[5].MuxAddr2 = 0x71
	qsfp.Vdev[5].MuxValue2 = 0x20

	qsfp.Vdev[6].Bus = 0
	qsfp.Vdev[6].Addr = 0x50
	qsfp.Vdev[6].MuxBus = 0
	qsfp.Vdev[6].MuxAddr = 0x70
	qsfp.Vdev[6].MuxValue = 0x1
	qsfp.Vdev[6].MuxBus2 = 0
	qsfp.Vdev[6].MuxAddr2 = 0x71
	qsfp.Vdev[6].MuxValue2 = 0x40

	qsfp.Vdev[7].Bus = 0
	qsfp.Vdev[7].Addr = 0x50
	qsfp.Vdev[7].MuxBus = 0
	qsfp.Vdev[7].MuxAddr = 0x70
	qsfp.Vdev[7].MuxValue = 0x1
	qsfp.Vdev[7].MuxBus2 = 0
	qsfp.Vdev[7].MuxAddr2 = 0x71
	qsfp.Vdev[7].MuxValue2 = 0x80

	qsfp.Vdev[8].Bus = 0
	qsfp.Vdev[8].Addr = 0x50
	qsfp.Vdev[8].MuxBus = 0
	qsfp.Vdev[8].MuxAddr = 0x70
	qsfp.Vdev[8].MuxValue = 0x2
	qsfp.Vdev[8].MuxBus2 = 0
	qsfp.Vdev[8].MuxAddr2 = 0x71
	qsfp.Vdev[8].MuxValue2 = 0x1

	qsfp.Vdev[9].Bus = 0
	qsfp.Vdev[9].Addr = 0x50
	qsfp.Vdev[9].MuxBus = 0
	qsfp.Vdev[9].MuxAddr = 0x70
	qsfp.Vdev[9].MuxValue = 0x2
	qsfp.Vdev[9].MuxBus2 = 0
	qsfp.Vdev[9].MuxAddr2 = 0x71
	qsfp.Vdev[9].MuxValue2 = 0x2

	qsfp.Vdev[10].Bus = 0
	qsfp.Vdev[10].Addr = 0x50
	qsfp.Vdev[10].MuxBus = 0
	qsfp.Vdev[10].MuxAddr = 0x70
	qsfp.Vdev[10].MuxValue = 0x2
	qsfp.Vdev[10].MuxBus2 = 0
	qsfp.Vdev[10].MuxAddr2 = 0x71
	qsfp.Vdev[10].MuxValue2 = 0x4

	qsfp.Vdev[11].Bus = 0
	qsfp.Vdev[11].Addr = 0x50
	qsfp.Vdev[11].MuxBus = 0
	qsfp.Vdev[11].MuxAddr = 0x70
	qsfp.Vdev[11].MuxValue = 0x2
	qsfp.Vdev[11].MuxBus2 = 0
	qsfp.Vdev[11].MuxAddr2 = 0x71
	qsfp.Vdev[11].MuxValue2 = 0x8

	qsfp.Vdev[12].Bus = 0
	qsfp.Vdev[12].Addr = 0x50
	qsfp.Vdev[12].MuxBus = 0
	qsfp.Vdev[12].MuxAddr = 0x70
	qsfp.Vdev[12].MuxValue = 0x2
	qsfp.Vdev[12].MuxBus2 = 0
	qsfp.Vdev[12].MuxAddr2 = 0x71
	qsfp.Vdev[12].MuxValue2 = 0x10

	qsfp.Vdev[13].Bus = 0
	qsfp.Vdev[13].Addr = 0x50
	qsfp.Vdev[13].MuxBus = 0
	qsfp.Vdev[13].MuxAddr = 0x70
	qsfp.Vdev[13].MuxValue = 0x2
	qsfp.Vdev[13].MuxBus2 = 0
	qsfp.Vdev[13].MuxAddr2 = 0x71
	qsfp.Vdev[13].MuxValue2 = 0x20

	qsfp.Vdev[14].Bus = 0
	qsfp.Vdev[14].Addr = 0x50
	qsfp.Vdev[14].MuxBus = 0
	qsfp.Vdev[14].MuxAddr = 0x70
	qsfp.Vdev[14].MuxValue = 0x2
	qsfp.Vdev[14].MuxBus2 = 0
	qsfp.Vdev[14].MuxAddr2 = 0x71
	qsfp.Vdev[14].MuxValue2 = 0x40

	qsfp.Vdev[15].Bus = 0
	qsfp.Vdev[15].Addr = 0x50
	qsfp.Vdev[15].MuxBus = 0
	qsfp.Vdev[15].MuxAddr = 0x70
	qsfp.Vdev[15].MuxValue = 0x2
	qsfp.Vdev[15].MuxBus2 = 0
	qsfp.Vdev[15].MuxAddr2 = 0x71
	qsfp.Vdev[15].MuxValue2 = 0x80

	qsfp.Vdev[16].Bus = 0
	qsfp.Vdev[16].Addr = 0x50
	qsfp.Vdev[16].MuxBus = 0
	qsfp.Vdev[16].MuxAddr = 0x70
	qsfp.Vdev[16].MuxValue = 0x4
	qsfp.Vdev[16].MuxBus2 = 0
	qsfp.Vdev[16].MuxAddr2 = 0x71
	qsfp.Vdev[16].MuxValue2 = 0x1

	qsfp.Vdev[17].Bus = 0
	qsfp.Vdev[17].Addr = 0x50
	qsfp.Vdev[17].MuxBus = 0
	qsfp.Vdev[17].MuxAddr = 0x70
	qsfp.Vdev[17].MuxValue = 0x4
	qsfp.Vdev[17].MuxBus2 = 0
	qsfp.Vdev[17].MuxAddr2 = 0x71
	qsfp.Vdev[17].MuxValue2 = 0x2

	qsfp.Vdev[18].Bus = 0
	qsfp.Vdev[18].Addr = 0x50
	qsfp.Vdev[18].MuxBus = 0
	qsfp.Vdev[18].MuxAddr = 0x70
	qsfp.Vdev[18].MuxValue = 0x4
	qsfp.Vdev[18].MuxBus2 = 0
	qsfp.Vdev[18].MuxAddr2 = 0x71
	qsfp.Vdev[18].MuxValue2 = 0x4

	qsfp.Vdev[19].Bus = 0
	qsfp.Vdev[19].Addr = 0x50
	qsfp.Vdev[19].MuxBus = 0
	qsfp.Vdev[19].MuxAddr = 0x70
	qsfp.Vdev[19].MuxValue = 0x4
	qsfp.Vdev[19].MuxBus2 = 0
	qsfp.Vdev[19].MuxAddr2 = 0x71
	qsfp.Vdev[19].MuxValue2 = 0x8

	qsfp.Vdev[20].Bus = 0
	qsfp.Vdev[20].Addr = 0x50
	qsfp.Vdev[20].MuxBus = 0
	qsfp.Vdev[20].MuxAddr = 0x70
	qsfp.Vdev[20].MuxValue = 0x4
	qsfp.Vdev[20].MuxBus2 = 0
	qsfp.Vdev[20].MuxAddr2 = 0x71
	qsfp.Vdev[20].MuxValue2 = 0x10

	qsfp.Vdev[21].Bus = 0
	qsfp.Vdev[21].Addr = 0x50
	qsfp.Vdev[21].MuxBus = 0
	qsfp.Vdev[21].MuxAddr = 0x70
	qsfp.Vdev[21].MuxValue = 0x4
	qsfp.Vdev[21].MuxBus2 = 0
	qsfp.Vdev[21].MuxAddr2 = 0x71
	qsfp.Vdev[21].MuxValue2 = 0x20

	qsfp.Vdev[22].Bus = 0
	qsfp.Vdev[22].Addr = 0x50
	qsfp.Vdev[22].MuxBus = 0
	qsfp.Vdev[22].MuxAddr = 0x70
	qsfp.Vdev[22].MuxValue = 0x4
	qsfp.Vdev[22].MuxBus2 = 0
	qsfp.Vdev[22].MuxAddr2 = 0x71
	qsfp.Vdev[22].MuxValue2 = 0x40

	qsfp.Vdev[23].Bus = 0
	qsfp.Vdev[23].Addr = 0x50
	qsfp.Vdev[23].MuxBus = 0
	qsfp.Vdev[23].MuxAddr = 0x70
	qsfp.Vdev[23].MuxValue = 0x4
	qsfp.Vdev[23].MuxBus2 = 0
	qsfp.Vdev[23].MuxAddr2 = 0x71
	qsfp.Vdev[23].MuxValue2 = 0x80

	qsfp.Vdev[24].Bus = 0
	qsfp.Vdev[24].Addr = 0x50
	qsfp.Vdev[24].MuxBus = 0
	qsfp.Vdev[24].MuxAddr = 0x70
	qsfp.Vdev[24].MuxValue = 0x8
	qsfp.Vdev[24].MuxBus2 = 0
	qsfp.Vdev[24].MuxAddr2 = 0x71
	qsfp.Vdev[24].MuxValue2 = 0x1

	qsfp.Vdev[25].Bus = 0
	qsfp.Vdev[25].Addr = 0x50
	qsfp.Vdev[25].MuxBus = 0
	qsfp.Vdev[25].MuxAddr = 0x70
	qsfp.Vdev[25].MuxValue = 0x8
	qsfp.Vdev[25].MuxBus2 = 0
	qsfp.Vdev[25].MuxAddr2 = 0x71
	qsfp.Vdev[25].MuxValue2 = 0x2

	qsfp.Vdev[26].Bus = 0
	qsfp.Vdev[26].Addr = 0x50
	qsfp.Vdev[26].MuxBus = 0
	qsfp.Vdev[26].MuxAddr = 0x70
	qsfp.Vdev[26].MuxValue = 0x8
	qsfp.Vdev[26].MuxBus2 = 0
	qsfp.Vdev[26].MuxAddr2 = 0x71
	qsfp.Vdev[26].MuxValue2 = 0x4

	qsfp.Vdev[27].Bus = 0
	qsfp.Vdev[27].Addr = 0x50
	qsfp.Vdev[27].MuxBus = 0
	qsfp.Vdev[27].MuxAddr = 0x70
	qsfp.Vdev[27].MuxValue = 0x8
	qsfp.Vdev[27].MuxBus2 = 0
	qsfp.Vdev[27].MuxAddr2 = 0x71
	qsfp.Vdev[27].MuxValue2 = 0x8

	qsfp.Vdev[28].Bus = 0
	qsfp.Vdev[28].Addr = 0x50
	qsfp.Vdev[28].MuxBus = 0
	qsfp.Vdev[28].MuxAddr = 0x70
	qsfp.Vdev[28].MuxValue = 0x8
	qsfp.Vdev[28].MuxBus2 = 0
	qsfp.Vdev[28].MuxAddr2 = 0x71
	qsfp.Vdev[28].MuxValue2 = 0x10

	qsfp.Vdev[29].Bus = 0
	qsfp.Vdev[29].Addr = 0x50
	qsfp.Vdev[29].MuxBus = 0
	qsfp.Vdev[29].MuxAddr = 0x70
	qsfp.Vdev[29].MuxValue = 0x8
	qsfp.Vdev[29].MuxBus2 = 0
	qsfp.Vdev[29].MuxAddr2 = 0x71
	qsfp.Vdev[29].MuxValue2 = 0x20

	qsfp.Vdev[30].Bus = 0
	qsfp.Vdev[30].Addr = 0x50
	qsfp.Vdev[30].MuxBus = 0
	qsfp.Vdev[30].MuxAddr = 0x70
	qsfp.Vdev[30].MuxValue = 0x8
	qsfp.Vdev[30].MuxBus2 = 0
	qsfp.Vdev[30].MuxAddr2 = 0x71
	qsfp.Vdev[30].MuxValue2 = 0x40

	qsfp.Vdev[31].Bus = 0
	qsfp.Vdev[31].Addr = 0x50
	qsfp.Vdev[31].MuxBus = 0
	qsfp.Vdev[31].MuxAddr = 0x70
	qsfp.Vdev[31].MuxValue = 0x8
	qsfp.Vdev[31].MuxBus2 = 0
	qsfp.Vdev[31].MuxAddr2 = 0x71
	qsfp.Vdev[31].MuxValue2 = 0x80

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
