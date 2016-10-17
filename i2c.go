// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package i2c provides cli command to access i2c devices.
package i2c

import (
	"fmt"

	"time"

	"github.com/platinasystems/goes/i2c"
	"github.com/platinasystems/oops"
)

type i2c_ struct{ oops.Id }

var I2c = &i2c_{"i2c"}

func (*i2c_) Usage() string {
	return "i2c ['EEPROM'] BUS.ADDR[.BEGIN][-END][/8][/16] [VALUE] [WRITE-DELAY-IN-SEC]"
}

func (p *i2c_) Main(args ...string) {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	var (
		bus        i2c.Bus
		sd         i2c.SMBusData
		b, a, d, w uint8
		cs         [2]uint8
	)

	if n := len(args); n == 0 {
		p.Panic("BUS.ADDR.REG: missing")
	} else if n > 3 {
		p.Panic(args[3:], ": unexpected")
	}

	eeprom := 0
	if args[0] == "EEPROM" {
		eeprom, args = 1, args[1:]
	}

	dValid := len(args) > 1

	nc := 2
	w = 0
	_, err := fmt.Sscanf(args[0], "%x.%x.%x-%x/%d", &b, &a, &cs[0], &cs[1], &w)
	if err != nil {
		_, err = fmt.Sscanf(args[0], "%x.%x.%x-%x", &b, &a, &cs[0], &cs[1])
		if err != nil {
			nc = 1
			_, err = fmt.Sscanf(args[0], "%x.%x.%x/%d", &b, &a, &cs[0], &w)
			if err != nil {
				_, err = fmt.Sscanf(args[0], "%x.%x.%x", &b, &a, &cs[0])
				if err != nil {
					nc = 0
					_, err = fmt.Sscanf(args[0], "%x.%x/%d", &b, &a, &w)
					if err != nil {
						_, err = fmt.Sscanf(args[0], "%x.%x", &b, &a)
					}
				}
			}
		}
	}
	if err != nil {
		p.Panic(args[0], ": invalid BUS.ADDR[.REG]: ", err)
	}
	if w != 0 && w != 8 && w != 16 {
		p.Panic(w, ": invalid R/W width: ", w)
	}

	if dValid {
		_, err = fmt.Sscanf(args[1], "%x", &d)
		if err != nil {
			p.Panic(args[1], ": invalid value: ", err)
		}
	}

	writeDelay := float64(0)
	if len(args) > 2 {
		s := args[2]
		_, err := fmt.Sscanf(s, "%f", &writeDelay)
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

	c := uint8(0)
	op := i2c.ByteData
	if eeprom == 1 {
		sd[0] = cs[0]
		err = bus.Do(i2c.Write, c, op, &sd)
		if err != nil {
			p.Panic(err)
		}
	}

	op = i2c.ByteData
	if nc == 0 || eeprom == 1 || w == 8 {
		op = i2c.Byte
	}
	if w == 16 {
		op = i2c.WordData
	}

	rw := i2c.Read
	if dValid {
		rw = i2c.Write
		sd[0] = d
	}

	c = cs[0]
	if nc < 2 {
		err = bus.Do(rw, c, op, &sd)
		if err != nil {
			p.Panic(err)
		}
		if w == 16 {
			fmt.Printf("%x.%02x.%02x = %02x\n", b, a, c, uint16(sd[1])<<8|uint16(sd[0]))
			return
		} else {
			fmt.Printf("%x.%02x.%02x = %02x\n", b, a, c, sd[0])
			return
		}
	}

	s := ""
	count := 0
	ascii := ""
	for {
		err = bus.Do(rw, c, op, &sd)
		if err != nil {
			p.Panic(err)
		}
		if count == 0 {
			s += fmt.Sprintf("%02x: ", c)
		}
		if w == 16 {
			s += fmt.Sprintf("%02x ", (uint16(sd[1])<<8 | uint16(sd[0])))
		} else {
			s += fmt.Sprintf("%02x ", sd[0])
		}
		if sd[0] < 0x7e && sd[0] > 0x1f {
			ascii += fmt.Sprintf("%c", sd[0])
		} else {
			ascii += "."
		}
		if c == cs[1] {
			break
		}
		c++
		count++
		if count == 16 {
			count = 0
			s += "   "
			s += ascii
			s += "\n"
			ascii = ""
		}
		if rw == i2c.Write && writeDelay > 0 {
			time.Sleep(time.Second * time.Duration(writeDelay))
		}
	}
	fmt.Println(s)
}
