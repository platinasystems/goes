// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/platina/mk2/lc1-bmc/ledgpiod"
	"github.com/platinasystems/go/internal/redis"
)

func init() { ledgpiod.Init = ledgpiodInit }

func ledgpiodInit() {

	ledgpiod.VpageByKey = map[string]uint8{

		"system.fan_direction": 0,
	}

	ver := 0
	ledgpiod.Vdev.Bus = 0
	ledgpiod.Vdev.Addr = 0x0
	ledgpiod.Vdev.MuxBus = 0x0
	ledgpiod.Vdev.MuxAddr = 0x76
	ledgpiod.Vdev.MuxValue = 0x2
	s, _ := redis.Hget(redis.DefaultHash, "eeprom.DeviceVersion")
	_, _ = fmt.Sscan(s, &ver)
	switch ver {
	case 0xff:
		ledgpiod.Vdev.Addr = 0x22
	case 0x00:
		ledgpiod.Vdev.Addr = 0x22
	default:
		ledgpiod.Vdev.Addr = 0x75
	}

	ledgpiod.WrRegDv["ledgpiod"] = "ledgpiod"
	ledgpiod.WrRegFn["ledgpiod.example"] = "example"
	ledgpiod.WrRegRng["ledgpiod.example"] = []string{"true", "false"}
}
