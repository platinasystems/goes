// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package i2c provides cli command to access i2c devices.
package i2c

import (
	"fmt"
	"time"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/i2c"
)

type Command struct{}

func (Command) String() string { return "i2c" }

func (Command) Usage() string {
	return `
i2c [EEPROM][BLOCK] BUS.ADDR[.BEGIN][-END, -CNT][/8][/16] [VALUE] [WR-DELAY-SEC]
`
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "read/write I2C bus devices",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Read/write I2C bus devices.

	Examples:
	    i2c 0.76.0 80          writes a 0x80
	    i2c 0.2f.1f            reads device 0x2f, register 0x1f
	    i2c 0.2f.1f-20         reads two bytes
	    i2c EEPROM 0.55.0-30   reads 0x0-0x30 from EEPROM
	    i2c BLOCK 1.58.99      reads upto 32 bytes from BLOCK at reg 0x99
	    i2c BLOCK 1.58.99-10   reads upto 0x10 bytes from BLOCK at reg 0x99
            i2c 0.76/8             force reads at 8-bits
            i2c 0.76.0/8           force reads at 8-bits
	    i2c 0.55.0-30/8        reads 0x0-0x30 8-bits at a time
	    i2c 0.55.0-30/16       reads 0x0-0x30 16-bits at a time`,
	}
}

func (Command) Main(args ...string) error {
	var (
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
	block := 0
	if args[0] == "EEPROM" {
		eeprom, args = 1, args[1:]
	}
	if args[0] == "BLOCK" {
		block, args = 1, args[1:]
	}
	if args[0] == "STOP" {
		sd[0] = 0
		j[0] = I{true, i2c.Write, 0, 0, sd, int(0x99), int(1), 0}
		err := DoI2cRpc()
		if err != nil {
			return err
		}
		return nil
	}
	if args[0] == "START" {
		sd[0] = 0
		j[0] = I{true, i2c.Write, 0, 0, sd, int(0x99), int(0), 0}
		err := DoI2cRpc()
		if err != nil {
			return err
		}
		return nil
	}
	if args[0] == "READ" {
		sd[0] = 0
		j[0] = I{true, i2c.Write, 0, 0, sd, int(0x98), int(0), 0}
		err := DoI2cRpc()
		if err != nil {
			return err
		}
		fmt.Printf("Stop polling bit is 0x%x\n", s[0].D[0])
		return nil
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

	c := uint8(0)
	op := i2c.ByteData

	if eeprom == 1 {
		sd[0] = cs[0]
		j[0] = I{true, i2c.Write, c, op, sd, int(b), int(a), 0}
		err := DoI2cRpc()
		if err != nil {
			return err
		}
	}
	if block == 1 {
		op = i2c.I2CBlockData
		if nc < 2 {
			sd[0] = 32 + 1
		} else {
			sd[0] = cs[1] + 1
			if sd[0] > 31 {
				sd[0] = 32 + 1
			}
			nc = 1
		}
	}

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
		j[0] = I{true, rw, c, op, sd, int(b), int(a), 0}
		err := DoI2cRpc()
		if err != nil {
			return err
		}
		if w == 16 {
			fmt.Printf("%x.%02x.%02x = %02x\n", b, a, c, uint16(s[0].D[1])<<8|uint16(s[0].D[0]))
			return nil
		} else {
			if block == 0 {
				fmt.Printf("%x.%02x.%02x = %02x\n", b, a, c, s[0].D[0])
			}
			if block == 1 {
				k := 1
				t := ""
				count := 0
				ascii := ""
				for {
					if count == 0 {
						t += fmt.Sprintf("%02x: ", c)
					}
					t += fmt.Sprintf("%02x ", s[0].D[k])
					if s[0].D[k] < 0x7e && s[0].D[k] > 0x1f {
						ascii += fmt.Sprintf("%c", s[0].D[k])
					} else {
						ascii += "."
					}
					count++
					k++
					if count == 8 || k >= int(sd[0]) {
						for z := count; z < 9; z++ {
							t += "   "
						}
						count = 0
						t += ascii
						t += "\n"
						ascii = ""
					}
					if k >= int(sd[0]) {
						break
					}
				}
				fmt.Println(t)
			}
			return nil
		}
	}

	t := ""
	count := 0
	ascii := ""
	for {
		j[0] = I{true, rw, c, op, sd, int(b), int(a), 0}
		err := DoI2cRpc()
		if err != nil {
			return err
		}
		if count == 0 {
			t += fmt.Sprintf("%02x: ", c)
		}
		if w == 16 {
			t += fmt.Sprintf("%02x ", (uint16(s[0].D[1])<<8 | uint16(s[0].D[0])))
		} else {
			t += fmt.Sprintf("%02x ", s[0].D[0])
		}
		if s[0].D[0] < 0x7e && s[0].D[0] > 0x1f {
			ascii += fmt.Sprintf("%c", s[0].D[0])
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
			t += "   "
			t += ascii
			t += "\n"
			ascii = ""
		}
		if rw == i2c.Write && writeDelay > 0 {
			time.Sleep(time.Second * time.Duration(writeDelay))
		}
	}
	fmt.Println(t)
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

func toByte(a byte) byte {
	b := a - 0x30
	if b > 9 {
		b = b - 7
	}
	return b
}
func hexToByte(s string) (int, []byte) {
	arr := make([]byte, 40)
	m := 0
	for k, v := range []byte(s) {
		ak := (byte(v))
		if k == (k>>1)*2 {
			arr[m] = toByte(ak) << 4
		} else {
			arr[m] += toByte(ak)
			m++
		}
	}
	return m, arr
}
