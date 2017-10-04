// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"time"

	"github.com/platinasystems/go/goes/cmd/eeprom/platina_eeprom"
	"github.com/platinasystems/go/goes/cmd/redisd"
)

func init() {
	redisd.Init = func() {
		platina_eeprom.Config(
			platina_eeprom.BusIndex(0),
			platina_eeprom.BusAddress(0x55),
			platina_eeprom.BusDelay(10*time.Millisecond),
			platina_eeprom.MinMacs(2),
			platina_eeprom.OUI([3]byte{0x02, 0x46, 0x8a}),
		)
		redisd.Machine = "platina-mk2-mc1-bmc"
		redisd.Devs = []string{"lo", "eth0"}
		redisd.Hook = platina_eeprom.RedisdHook
	}
}
