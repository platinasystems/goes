// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package i2c provides cli command to access i2c devices.
package i2c

import (
	"fmt"
	"time"

	"github.com/platinasystems/go/goes/internal/i2c"
)

const Name = "i2c"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string {
	return Name + " ['EEPROM'] BUS.ADDR[.BEGIN][-END][/8][/16] [VALUE] [WRITE-DELAY-IN-SEC]"
}

func (cmd) Main(args ...string) error {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	var (
		bus        i2c.Bus
		sd         i2c.SMBusData
		b, a, d, w uint8
		cs         [2]uint8
	)

	if n := len(args); n == 0 {
		return fmt.Errorf("BUS.ADDR.REG: missing")
	} else if n > 3 {
		return fmt.Errorf("%v: unexpected", args[3:])
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
		return fmt.Errorf("%s: invalid BUS.ADDR[.REG]: %v", args[0], err)
	}
	if w != 0 && w != 8 && w != 16 {
		return fmt.Errorf("%v: invalid R/W width")
	}

	if dValid {
		_, err = fmt.Sscanf(args[1], "%x", &d)
		if err != nil {
			return fmt.Errorf("%s: invalid: %v", args[1], err)
		}
	}

	writeDelay := float64(0)
	if len(args) > 2 {
		s := args[2]
		_, err := fmt.Sscanf(s, "%f", &writeDelay)
		if err != nil {
			return fmt.Errorf("%s: invalid: %v", s, err)
		}
	}

	err = bus.Open(int(b))
	if err != nil {
		return err
	}
	defer bus.Close()

	err = bus.ForceSlaveAddress(int(a))
	if err != nil {
		return err
	}

	c := uint8(0)
	op := i2c.ByteData
	if eeprom == 1 {
		sd[0] = cs[0]
		err = bus.Do(i2c.Write, c, op, &sd)
		if err != nil {
			return err
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
			return err
		}
		if w == 16 {
			fmt.Printf("%x.%02x.%02x = %02x\n", b, a, c, uint16(sd[1])<<8|uint16(sd[0]))
			return nil
		} else {
			fmt.Printf("%x.%02x.%02x = %02x\n", b, a, c, sd[0])
			return nil
		}
	}

	s := ""
	count := 0
	ascii := ""
	for {
		err = bus.Do(rw, c, op, &sd)
		if err != nil {
			return err
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
	return nil
}

func ReadByte(b uint8, a uint8, c uint8) (uint8, error) {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock() //FIX THIS EVERYWHERE
	var (
		bus i2c.Bus
		sd  i2c.SMBusData
	)
	err := bus.Open(int(b))
	if err != nil {
		return 0, err
	}
	defer bus.Close()
	err = bus.ForceSlaveAddress(int(a))
	if err != nil {
		return 0, err
	}
	rw := i2c.Read
	op := i2c.ByteData
	err = bus.Do(rw, c, op, &sd)
	if err != nil {
		return 0, err
	}
	return sd[0], nil
}

func WriteByte(b uint8, a uint8, c uint8, v uint8) error {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()
	var (
		bus i2c.Bus
		sd  i2c.SMBusData
	)
	err := bus.Open(int(b))
	if err != nil {
		return err
	}
	defer bus.Close()
	err = bus.ForceSlaveAddress(int(a))
	if err != nil {
		return err
	}
	rw := i2c.Write
	op := i2c.ByteData
	sd[0] = v
	err = bus.Do(rw, c, op, &sd)
	if err != nil {
		return err
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "read/write I2C bus devices",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	i2c - Read/write I2C bus devices

SYNOPSIS
	i2c

DESCRIPTION
	Read/write I2C bus devices.

	Examples:
	    i2c 0.76.0 80          writes a 0x80
	    i2c 0.2f.1f            reads device 0x2f, register 0x1f
	    i2c 0.2f.1f-20         reads two bytes
	    i2c EEPROM 0.55.0-30   reads 0x0-0x30 from EEPROM
            i2c 0.76/8             force reads at 8-bits
            i2c 0.76.0/8           force reads at 8-bits
	    i2c 0.55.0-30/8        reads 0x0-0x30 8-bits at a time
	    i2c 0.55.0-30/16       reads 0x0-0x30 16-bits at a time`,
	}
}
