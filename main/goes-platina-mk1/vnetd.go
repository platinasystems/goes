// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"

	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/goes/cmd/vnetd"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/sriovs"
	"github.com/platinasystems/go/vnet"
	fe1_platform "github.com/platinasystems/go/vnet/platforms/fe1"
	mk1 "github.com/platinasystems/go/vnet/platforms/mk1"
)

type mk1Main struct {
	fe1_platform.Platform
}

var defaultMk1 = &mk1Main{}

func init() { vnetd.Hook = defaultMk1.vnetdHook }

func (p *mk1Main) vnetdHook(init func(), v *vnet.Vnet) error {
	p.Init = init

	s, err := redis.Hget(redis.DefaultHash, "eeprom.DeviceVersion")
	if err != nil {
		return err
	}
	if _, err = fmt.Sscan(s, &p.Version); err != nil {
		return err
	}
	s, err = redis.Hget(redis.DefaultHash, "eeprom.NEthernetAddress")
	if err != nil {
		return err
	}
	if _, err = fmt.Sscan(s, &p.NEthernetAddress); err != nil {
		return err
	}
	s, err = redis.Hget(redis.DefaultHash, "eeprom.BaseEthernetAddress")
	if err != nil {
		return err
	}
	input := new(parse.Input)
	input.SetString(s)
	p.BaseEthernetAddress.Parse(input)

	fns, err := sriovs.NumvfsFns()
	p.SriovMode = err == nil && len(fns) > 0

	vnetd.UnixInterfacesOnly = !p.SriovMode
	vnetd.GdbWait = gdbwait

	// Default to using MSI versus INTX for switch chip.
	p.EnableMsiInterrupt = true

	if err = mk1.PlatformInit(v, &p.Platform); err != nil {
		return err
	}

	return nil
}
