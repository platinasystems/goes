// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"

	"github.com/platinasystems/go/internal/goes/cmd/platina/mk1/bmc/ledgpio"
	"github.com/platinasystems/go/internal/redis"
)

func init() { ledgpio.Init = ledgpioInit }

func ledgpioInit() {
	ver := 0
	ledgpio.Vdev.Bus = 0
	ledgpio.Vdev.Addr = 0x0
	ledgpio.Vdev.MuxBus = 0x0
	ledgpio.Vdev.MuxAddr = 0x76
	ledgpio.Vdev.MuxValue = 0x2
	s, _ := redis.Hget(redis.DefaultHash, "eeprom.DeviceVersion")
	_, _ = fmt.Sscan(s, &ver)
	switch ver {
	case 0xff:
		ledgpio.Vdev.Addr = 0x22
	case 0x00:
		ledgpio.Vdev.Addr = 0x22
	default:
		ledgpio.Vdev.Addr = 0x75
	}
}
