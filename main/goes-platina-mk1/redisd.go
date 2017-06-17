// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"time"

	"github.com/platinasystems/go/goes/cmd/eeprom/platina_eeprom"
	"github.com/platinasystems/go/goes/cmd/redisd"
	"github.com/platinasystems/go/internal/redis/publisher"
)

func init() {
	redisd.Init = func() {
		platina_eeprom.Config(
			platina_eeprom.BusIndex(0),
			platina_eeprom.BusAddress(0x51),
			platina_eeprom.BusDelay(10*time.Millisecond),
			platina_eeprom.MinMacs(132),
			platina_eeprom.OUI([3]byte{0x02, 0x46, 0x8a}),
		)
		redisd.Machine = "platina-mk1"
		redisd.Devs = []string{"lo", "eth0"}
		redisd.Hook = func(pub *publisher.Publisher) {
			pub.Print("sriov.numvfs: 63")
			platina_eeprom.RedisdHook(pub)
		}
	}
}
