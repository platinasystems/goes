// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import "github.com/platinasystems/go/goes/cmd/w83795d"

func init() { w83795d.Init = w83795dInit }

func w83795dInit() {
	w83795d.Vdev.Bus = 0
	w83795d.Vdev.Addr = 0x2f
	w83795d.Vdev.MuxBus = 0
	w83795d.Vdev.MuxAddr = 0x76
	w83795d.Vdev.MuxValue = 0x80

	w83795d.VpageByKey = map[string]uint8{
		"fan_tray.1.1.speed.units.rpm": 1,
		"fan_tray.1.2.speed.units.rpm": 2,
		"fan_tray.2.1.speed.units.rpm": 3,
		"fan_tray.2.2.speed.units.rpm": 4,
		"fan_tray.3.1.speed.units.rpm": 5,
		"fan_tray.3.2.speed.units.rpm": 6,
		"fan_tray.4.1.speed.units.rpm": 7,
		"fan_tray.4.2.speed.units.rpm": 8,
		"fan_tray.speed":               0,
		"fan_tray.duty":                0,
		"hwmon.front.temp.units.C":     0,
		"hwmon.rear.temp.units.C":      0,
		"host.temp.units.C":            0,
		"host.temp.target.units.C":     0,
	}

	w83795d.WrRegDv["fan_tray"] = "fan_tray"
	w83795d.WrRegFn["fan_tray.example"] = "example"
	w83795d.WrRegFn["fan_tray.speed"] = "speed"
	w83795d.WrRegDv["host"] = "host"
	w83795d.WrRegFn["host.temp.units.C"] = "temp.units.C"
	w83795d.WrRegFn["host.temp.target.units.C"] = "temp.target.units.C"
	w83795d.WrRegRng["fan_tray.speed"] = []string{"low", "med", "high", "auto"}
	w83795d.WrRegRng["w83795d.example"] = []string{"true", "false"}
}
