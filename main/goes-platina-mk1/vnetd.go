// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/goes/cmd/ip"
	"github.com/platinasystems/go/goes/cmd/vnetd"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/platforms/fe1"
	"github.com/platinasystems/go/vnet/platforms/mk1"
)

type mk1Main struct {
	fe1.Platform
}

func vnetdInit() {
	p := new(mk1Main)
	vnetd.Hook = p.vnetdHook
	vnetd.CloseHook = p.stopHook
}

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

	fi, err := os.Stat("/sys/bus/pci/drivers/ixgbe")
	p.KernelIxgbe = err == nil && fi.IsDir()

	vnetd.UnixInterfacesOnly = !p.KernelIxgbe

	// Default to using MSI versus INTX for switch chip.
	p.EnableMsiInterrupt = true

	if err = mk1.PlatformInit(v, &p.Platform); err != nil {
		return err
	}

	return nil
}

func (p *mk1Main) stopHook(i *vnetd.Info, v *vnet.Vnet) error {
	if p.KernelIxgbe {
		return mk1.PlatformExit(v, &p.Platform)
	} else {
		// FIXME why isn't this done in mk1.PlatformExit?
		interfaces, err := net.Interfaces()
		if err != nil {
			return err
		}
		for _, dev := range interfaces {
			if strings.HasPrefix(dev.Name, "eth-") ||
				dev.Name == "vnet" {
				args := []string{"link", "delete", dev.Name}
				err = ip.New().Main(args...)
				if err != nil {
					fmt.Println("write err", err)
					return err
				}
			}
		}
		return nil
	}
}
