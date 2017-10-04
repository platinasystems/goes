// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"time"

	"github.com/platinasystems/go/goes/cmd/start"
	"github.com/platinasystems/go/internal/gpio"
	"github.com/platinasystems/go/internal/log"
)

func init() {
	start.ConfGpioHook = func() error {
		gpio.Init()
		pin, found := gpio.Pins["QSPI_MUX_SEL"]
		if found {
			r, _ := pin.Value()
			if r {
				log.Print("Booted from QSPI1")
			} else {
				log.Print("Booted from QSPI0")
			}

		}

		for name, pin := range gpio.Pins {
			err := pin.SetDirection()
			if err != nil {
				fmt.Printf("%s: %v\n", name, err)
			}
		}
		pin, found = gpio.Pins["LOCAL_I2C_RESET_L"]
		if found {
			pin.SetValue(false)
			time.Sleep(1 * time.Microsecond)
			pin.SetValue(true)
		}

		pin, found = gpio.Pins["FRU_I2C_RESET_L"]
		if found {
			pin.SetValue(false)
			time.Sleep(1 * time.Microsecond)
			pin.SetValue(true)
		}
		return nil
	}
}
