// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package i2c provides cli command to access i2c devices.
package i2c

import (
	"fmt"

	"time"

	"github.com/platinasystems/i2c"
	"github.com/platinasystems/oops"
)

type i2c_ struct{ oops.Id }

var I2c = &i2c_{"i2c"}

func (*i2c_) Usage() string {
	return "i2c BUS.ADDR[.REG_BEGIN][-REG_END] [OP] [VALUE] [WRITE-DELAY-IN-SEC]"
}

func (p *i2c_) Main(args ...string) {
	var (
		bus           i2c.Bus
		sd            i2c.SMBusData
		b, a, d, o, t uint8
		cs            [2]uint8
	)

	if n := len(args); n == 0 {
		p.Panic("BUS.ADDR.REG: missing")
	} else if n > 4 {
		p.Panic(args[4:], ": unexpected")
	}

	oValid := len(args) > 1
	dValid := len(args) > 2

	nc := 2
	_, err := fmt.Sscanf(args[0], "%x.%x.%x-%x", &b, &a, &cs[0], &cs[1])
	if err != nil {
		nc = 1
		_, err = fmt.Sscanf(args[0], "%x.%x.%x", &b, &a, &cs[0])
		if err != nil {
			nc = 0
			_, err = fmt.Sscanf(args[0], "%x.%x", &b, &a)
		}
	}
	if err != nil {
		p.Panic(args[0], ": invalid BUS.ADDR[.REG]: ", err)
	}

	if oValid {
		_, err = fmt.Sscanf(args[1], "%x", &o)
		if err != nil {
			p.Panic(args[1], ": invalid value: ", err)
		}
	}

	if dValid {
		_, err = fmt.Sscanf(args[2], "%x", &d)
		if err != nil {
			p.Panic(args[2], ": invalid value: ", err)
		}
	}

	if len(args) > 3 {
		s := args[3]
		_, err := fmt.Sscanf(s, "%x", &t)
		if err != nil {
			p.Panic(s, ": invalid delay: ", err)
		}
	}

	err = bus.Open(int(b))
	if err != nil {
		p.Panic(err)
	}
	defer bus.Close()

	err = bus.ForceSlaveAddress(int(a))
	if err != nil {
		p.Panic(err)
	}

	op := i2c.ByteData
	if nc == 0 {
		op = i2c.Byte
	}
	if oValid {
		op = (i2c.SMBusSize)(o)
	}

	rw := i2c.Read
	if dValid {
		rw = i2c.Write
		sd[0] = d
	}

	if nc < 2 {
		cs[1] = cs[0]
	}
	c := cs[0]
	for {
		err = bus.Do(rw, c, op, &sd)
		if err != nil {
			p.Panic(err)
		}
		fmt.Printf("%x.%02x.%02x = %02x\n", b, a, c, sd[0])
		if c == cs[1] {
			break
		}
		c++
		if rw == i2c.Write && t > 0 {
			time.Sleep(time.Second * time.Duration(t))
		}
	}
}
