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

	To set the address for a RANDOM READ from an I2C EEPROM device,
	use OP=2 and specify a VALUE for the address.
	Example:  i2c 0.55.0 2 40

	To read from an I2C EEPROM I2C device, use OP=1 without VALUE.
	Example:  i2c 0.55.0 1    The address will auto-increment.`
}
