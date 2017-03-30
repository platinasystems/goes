// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This is an example Baseboard Management Controller.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/platinasystems/go/internal/goes/cmd/eeprom/platina_eeprom"
)

func main() {
	g := mkgoes()
	platina_eeprom.Config(
		platina_eeprom.BusIndex(0),
		platina_eeprom.BusAddress(0x55),
		platina_eeprom.BusDelay(10*time.Millisecond),
		platina_eeprom.MinMacs(2),
		platina_eeprom.OUI([3]byte{0x02, 0x46, 0x8a}),
	)
	if err := Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := g.Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
