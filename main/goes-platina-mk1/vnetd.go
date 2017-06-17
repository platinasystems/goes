// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"

	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/goes/cmd/vnetd"
	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/sriovs"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/plugin/fe1"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"
)

func init() {
	vnetd.Hook = func(i *vnetd.Info, v *vnet.Vnet) error {
		var ver int
		var nmacs uint32
		var basea ethernet.Address
		s, err := redis.Hget(redis.DefaultHash,
			"eeprom.DeviceVersion")
		if err != nil {
			return err
		}
		if _, err = fmt.Sscan(s, &ver); err != nil {
			return err
		}
		s, err = redis.Hget(redis.DefaultHash,
			"eeprom.NEthernetAddress")
		if err != nil {
			return err
		}
		if _, err = fmt.Sscan(s, &nmacs); err != nil {
			return err
		}
		s, err = redis.Hget(redis.DefaultHash,
			"eeprom.BaseEthernetAddress")
		if err != nil {
			return err
		}
		input := new(parse.Input)
		input.SetString(s)
		basea.Parse(input)

		fns, err := sriovs.NumvfsFns()
		have_numvfs := err == nil && len(fns) > 0

		vnetd.UnixInterfacesOnly = !have_numvfs
		vnetd.GdbWait = gdbwait

		// Base packages.
		ethernet.Init(v)
		ip4.Init(v)
		ip6.Init(v)
		pg.Init(v) // vnet packet generator
		unix.Init(v)

		// Device drivers.
		fe1.Init(v)
		if !have_numvfs {
			ixge.Init(v, ixge.Config{DisableUnix: true, PuntNode: "fe1-single-tagged-punt"})
		} else if err = newSriovs(ver); err != nil {
			return err
		}

		fe1.AddPlatform(v, ver, nmacs, basea, i.Init, leden)

		return nil
	}
}

// MK1 board front panel port LED's require PCA9535 GPIO device
// configuration - to provide an output signal that allows LED
// operation.
func leden() (err error) {
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
