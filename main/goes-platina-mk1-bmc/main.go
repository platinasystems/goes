// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This is an example Baseboard Management Controller.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/platinasystems/go/internal/fdt"
	"github.com/platinasystems/go/internal/fdtgpio"
	"github.com/platinasystems/go/internal/goes/cmd/eeprom/platina_eeprom"
	"github.com/platinasystems/go/internal/goes/cmd/start"
	"github.com/platinasystems/go/internal/gpio"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/redis"
)

const UsrShareGoes = "/usr/share/goes"

func main() {
	gpio.File = "/boot/platina-mk1-bmc.dtb"
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
	start.ConfGpioHook = func() error {
		gpio.Aliases = make(gpio.GpioAliasMap)
		gpio.Pins = make(gpio.PinMap)
		if b, err := ioutil.ReadFile(gpio.File); err == nil {
			t := &fdt.Tree{Debug: false, IsLittleEndian: false}
			t.Parse(b)

			t.MatchNode("aliases", fdtgpio.GatherAliases)
			t.EachProperty("gpio-controller", "", fdtgpio.GatherPins)
		} else {
			return fmt.Errorf("%s: %v", gpio.File, err)
		}
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
		pin, found = gpio.Pins["FRU_I2C_MUX_RST_L"]
		if found {
			pin.SetValue(false)
			time.Sleep(1 * time.Microsecond)
			pin.SetValue(true)
		}

		pin, found = gpio.Pins["MAIN_I2C_MUX_RST_L"]
		if found {
			pin.SetValue(false)
			time.Sleep(1 * time.Microsecond)
			pin.SetValue(true)
		}
		redis.Hwait(redis.DefaultHash, "redis.ready", "true",
			10*time.Second)
		s, err := redis.Hget(redis.DefaultHash, "eeprom.DeviceVersion")
		if err != nil {
			log.Print(err)
			return err
		}
		ver := 0
		_, err = fmt.Sscan(s, &ver)
		if err != nil {
			log.Print(err)
			return err
		}
		f, err := os.Create("/tmp/ver")
		if err != nil {
			return err
		}
		defer f.Close()
		d2 := []byte{byte(ver), 10}
		_, err = f.Write(d2)
		if err != nil {
			return err
		}
		f.Sync()
		f.Close()
		return nil
	}
	if err := g.Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
