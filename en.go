// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build !noten

package i2c

func (*i2c_) Apropos() string {
	return "read/write I2C bus devices"
}

func (*i2c_) Man() string {
	return `NAME
	i2c - Read/write I2C bus devices

SYNOPSIS
	i2c

DESCRIPTION
	Read/write I2C bus devices.

	Examples:
	    i2c 0.76.0 80          writes a 0x80
	    i2c 0.2f.1f            reads device 0x2f, register 0x1f
	    i2c 0.2f.1f-20         reads two bytes
	    i2c EEPROM 0.55.0-30   reads 0x0-0x30 from EEPROM`

}
