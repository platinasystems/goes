// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/platina/mk1/bmc/ucd9090d"
	"github.com/platinasystems/go/internal/redis"
)

func ucd9090dInit() {
	ver := 0
	ucd9090d.Vdev.Bus = 0
	ucd9090d.Vdev.Addr = 0x0 //update after eeprom read
	ucd9090d.Vdev.MuxBus = 0
	ucd9090d.Vdev.MuxAddr = 0x76
	ucd9090d.Vdev.MuxValue = 0x01
	s, err := redis.Hget(redis.DefaultHash, "eeprom.DeviceVersion")
	if err != nil {
		ucd9090d.Vdev.Addr = 0x34
	} else {
		_, _ = fmt.Sscan(s, &ver)
		switch ver {
		case 0xff:
			ucd9090d.Vdev.Addr = 0x7e
		case 0x00:
			ucd9090d.Vdev.Addr = 0x7e
		default:
			ucd9090d.Vdev.Addr = 0x34
		}
	}
	ucd9090d.VpageByKey = map[string]uint8{
		"vmon.5v.sb.units.V":    1,
		"vmon.3v8.bmc.units.V":  2,
		"vmon.3v3.sys.units.V":  3,
		"vmon.3v3.bmc.units.V":  4,
		"vmon.3v3.sb.units.V":   5,
		"vmon.1v0.thc.units.V":  6,
		"vmon.1v8.sys.units.V":  7,
		"vmon.1v25.sys.units.V": 8,
		"vmon.1v2.ethx.units.V": 9,
		"vmon.1v0.tha.units.V":  10,
		"vmon.poweroff.events":  0,
	}

	ucd9090d.WrRegDv["vmon"] = "vmon"
	ucd9090d.WrRegFn["vmon.example"] = "example"
	ucd9090d.WrRegRng["vmon.example"] = []string{"1", "50"}

	ucd9090d.WrRegDv["watchdog"] = "watchdog"
	ucd9090d.WrRegFn["watchdog.enable"] = "watchdog.enable"
	ucd9090d.WrRegFn["watchdog.sequence"] = "watchdog.sequence"
	ucd9090d.WrRegFn["watchdog.timeout.units.seconds"] = "watchdog.timeout.units.seconds"

	ucd9090d.WrRegRng["watchdog.enable"] = []string{"false", "true"}
	ucd9090d.WrRegRng["watchdog.host.timeout.units.seconds"] = []string{"0", "3600"}
}
