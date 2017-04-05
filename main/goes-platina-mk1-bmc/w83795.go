// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import "github.com/platinasystems/go/internal/goes/cmd/w83795"

func init() { w83795.Init = w83795Init }

func w83795Init() {
	w83795.Vdev.Bus = 0
	w83795.Vdev.Addr = 0x2f
	w83795.Vdev.MuxBus = 0
	w83795.Vdev.MuxAddr = 0x76
	w83795.Vdev.MuxValue = 0x80

	w83795.VpageByKey = map[string]uint8{
		"fan_tray.1.1.speed.units.rpm": 1,
		"fan_tray.1.2.speed.units.rpm": 2,
		"fan_tray.2.1.speed.units.rpm": 3,
		"fan_tray.2.2.speed.units.rpm": 4,
		"fan_tray.3.1.speed.units.rpm": 5,
		"fan_tray.3.2.speed.units.rpm": 6,
		"fan_tray.4.1.speed.units.rpm": 7,
		"fan_tray.4.2.speed.units.rpm": 8,
		"fan_tray.speed":               1,
	}

	w83795.WrRegDv["fan_tray"] = "fan_tray"
	w83795.WrRegFn["fan_tray.example"] = "example"
	w83795.WrRegFn["fan_tray.speed"] = "speed"
}
