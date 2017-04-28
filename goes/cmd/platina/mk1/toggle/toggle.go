// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package toggle

import (
	"net/rpc"
	"time"

	"github.com/platinasystems/go/goes/cmd/i2c"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/gpio"
	ii2c "github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/log"
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

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }
func (cmd) Man() lang.Alt     { return man }
func (cmd) String() string    { return Name }
func (cmd) Usage() string     { return Usage }

var clientA *rpc.Client
var dialed int = 0
var j [MAXOPS]I
var s [MAXOPS]R
var i = I{false, ii2c.RW(0), 0, 0, b, 0, 0, 0}
var x int
var b = [34]byte{0}

const MAXOPS = 30

var sd ii2c.SMBusData

type I struct {
	InUse     bool
	RW        ii2c.RW
	RegOffset uint8
	BusSize   ii2c.SMBusSize
	Data      [34]byte
	Bus       int
	Addr      int
	Delay     int
}
type R struct {
	D [34]byte
	E error
}

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
		j[0] = I{true, ii2c.Write, 0, 0, sd, int(0x99), int(1), 0}
		err := DoI2cRpc()
		if err != nil {
			return err
		}
		pin, found := gpio.Pins["CPU_TO_MAIN_I2C_EN"]
		if found {
			pin.SetValue(true)
		}
		time.Sleep(10 * time.Millisecond)
		x, _ := i2c.ReadByte(0, 0x74, 2)
		i2c.WriteByte(0, 0x74, 0x2, x|0x20)
		x, _ = i2c.ReadByte(0, 0x74, 6)
		i2c.WriteByte(0, 0x74, 0x6, x|0x20)
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
		j[0] = I{true, ii2c.Write, 0, 0, sd, int(0x99), int(0), 0}
		err = DoI2cRpc()
		if err != nil {
			return err
		}
	} else {
		x, _ := i2c.ReadByte(0, 0x74, 2)
		i2c.WriteByte(0, 0x74, 0x2, x|0x20)
		x, _ = i2c.ReadByte(0, 0x74, 6)
		i2c.WriteByte(0, 0x74, 0x6, x^0x20)
	}
	return nil
}

func clearJ() {
	x = 0
	for k := 0; k < MAXOPS; k++ {
		j[k] = i
	}
}

func DoI2cRpc() error {
	if dialed == 0 {
		client, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1233")
		if err != nil {
			log.Print("dialing:", err)
			return err
		}
		clientA = client
		dialed = 1
		time.Sleep(time.Millisecond * time.Duration(50))
	}
	err := clientA.Call("I2cReq.ReadWrite", &j, &s)
	if err != nil {
		log.Print("i2cReq error:", err)
		dialed = 0
		clientA.Close()
		return err
	}
	clearJ()
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
