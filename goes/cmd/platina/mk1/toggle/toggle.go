// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package toggle

import (
	"strings"
	"time"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/gpio"
	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/redis"
)

const (
	Name    = "toggle"
	Apropos = "toggle console port between x86 and BMC"
	Usage   = "toggle SECONDS"
	Man     = `
DESCRIPTION
	The toggle command toggles the console port between x86 and BMC.`
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

const (
	i2cGpioAddr = 0x74
)

func uartToggle() {
	var dir0, out0 uint8
	i2c.Do(0, i2cGpioAddr,
		func(bus *i2c.Bus) (err error) {
			var d i2c.SMBusData
			reg := uint8(6)
			err = bus.Read(reg, i2c.WordData, &d)
			dir0 = d[0]
			return
		})
	i2c.Do(0, i2cGpioAddr,
		func(bus *i2c.Bus) (err error) {
			var d i2c.SMBusData
			reg := uint8(6)
			err = bus.Read(reg, i2c.WordData, &d)
			out0 = d[0]
			return
		})
	i2c.Do(0, i2cGpioAddr,
		func(bus *i2c.Bus) (err error) {
			var d i2c.SMBusData
			d[0] = out0 | 0x20
			reg := uint8(2)
			err = bus.Write(reg, i2c.ByteData, &d)
			return
		})
	i2c.Do(0, i2cGpioAddr,
		func(bus *i2c.Bus) (err error) {
			var d i2c.SMBusData
			d[0] = dir0 ^ 0x20
			reg := uint8(6)
			err = bus.Write(reg, i2c.ByteData, &d)
			return
		})

}

func (Command) Main(args ...string) error {

	var machineBmc bool

	m, _ := redis.Hget(redis.DefaultHash, "machine")
	if strings.Contains(m, "bmc") {
		machineBmc = true
	} else {
		machineBmc = false
	}

	if machineBmc {
		cmd.Init("gpio")
		pin, found := gpio.Pins["CPU_TO_MAIN_I2C_EN"]
		if found {
			pin.SetValue(true)
		}
		time.Sleep(10 * time.Millisecond)
		uartToggle()
		if found {
			pin.SetValue(false)
		}
		pin, found = gpio.Pins["FP_BTN_UARTSEL_EN_L"]
		if found {
			pin.SetValue(true)
		}
		time.Sleep(10 * time.Millisecond)
	} else {
		uartToggle()
	}
	return nil
}

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
