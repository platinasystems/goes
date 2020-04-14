// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package toggle

import (
	"sync"

	"github.com/platinasystems/goes/external/i2c"
	"github.com/platinasystems/goes/lang"
)

const i2cGpioAddr = 0x74

type Command struct {
	Init func()
	init sync.Once
}

func (*Command) String() string { return "toggle" }

func (*Command) Usage() string { return "toggle SECONDS" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "toggle console port between x86 and BMC",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	The toggle command toggles the console port between x86 and BMC.`,
	}
}

func (c *Command) Main(args ...string) error {
	if c.Init != nil {
		c.init.Do(c.Init)
	}

	uartToggle()

	return nil
}

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
