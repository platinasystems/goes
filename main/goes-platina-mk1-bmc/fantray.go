// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import "github.com/platinasystems/go/internal/goes/cmd/fantray"

func init() { fantray.Init = fantrayInit }

func fantrayInit() {
	fantray.Vdev.Bus = 1
	fantray.Vdev.Addr = 0x20
	fantray.Vdev.MuxBus = 1
	fantray.Vdev.MuxAddr = 0x72
	fantray.Vdev.MuxValue = 0x04

	fantray.VpageByKey = map[string]uint8{
		"fan_tray.1.status": 1,
		"fan_tray.2.status": 2,
		"fan_tray.3.status": 3,
		"fan_tray.4.status": 4,
	}

	fantray.WrRegFn["fan_tray.example"] = "example"
	fantray.WrRegFn["fan_tray.speed"] = "speed"
}
