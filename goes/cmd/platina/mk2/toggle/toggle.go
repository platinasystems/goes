// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package toggle

import (
	"time"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/gpio"
	"github.com/platinasystems/i2c"
	"github.com/platinasystems/redis"
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

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }
func (cmd) Man() lang.Alt     { return man }
func (cmd) String() string    { return Name }
func (cmd) Usage() string     { return Usage }

var sd i2c.SMBusData

func (cmd) Main(args ...string) error {

	var machineBmc bool

	m, _ := redis.Hget(redis.DefaultHash, "machine")
	if m == "platina-mk1-bmc" {
		machineBmc = true
	} else {
		machineBmc = false
	}

	if machineBmc {
		if len(gpio.Pins) == 0 {
			gpio.Init()
		}

		//i2c STOP
		sd[0] = 0
		j[0] = I{true, i2c.Write, 0, 0, sd, int(0x99), int(1), 0}
		err := DoI2cRpc()
		if err != nil {
			return err
		}
		pin, found := gpio.Pins["CPU_TO_MAIN_I2C_EN"]
		if found {
			pin.SetValue(true)
		}
		time.Sleep(10 * time.Millisecond)
		x, _ := ReadByte(0, 0x74, 2)
		WriteByte(0, 0x74, 0x2, x|0x20)
		x, _ = ReadByte(0, 0x74, 6)
		WriteByte(0, 0x74, 0x6, x|0x20)
		if found {
			pin.SetValue(false)
		}
		pin, found = gpio.Pins["FP_BTN_UARTSEL_EN_L"]
		if found {
			pin.SetValue(true)
		}
		time.Sleep(10 * time.Millisecond)

		//i2c START
		sd[0] = 0
		j[0] = I{true, i2c.Write, 0, 0, sd, int(0x99), int(0), 0}
		err = DoI2cRpc()
		if err != nil {
			return err
		}
	} else {
		x, _ := ReadByte(0, 0x74, 2)
		WriteByte(0, 0x74, 0x2, x|0x20)
		x, _ = ReadByte(0, 0x74, 6)
		WriteByte(0, 0x74, 0x6, x^0x20)
	}
	return nil
}

func ReadByte(b uint8, a uint8, c uint8) (uint8, error) {
	var (
		sd i2c.SMBusData
	)
	rw := i2c.Read
	op := i2c.ByteData
	j[0] = I{true, rw, c, op, sd, int(b), int(a), 0}
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	return s[0].D[0], nil
}

func WriteByte(b uint8, a uint8, c uint8, v uint8) error {
	var (
		sd i2c.SMBusData
	)
	rw := i2c.Write
	op := i2c.ByteData
	sd[0] = v
	j[0] = I{true, rw, c, op, sd, int(b), int(a), 0}
	err := DoI2cRpc()
	if err != nil {
		return err
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
