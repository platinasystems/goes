// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build !noten

package i2c

func (p *I2c) Apropos() string {
	return "read/write I2C bus devices"
}

func (p *I2c) Man() string {
	return `NAME
	i2c - Read/write I2C bus devices

SYNOPSIS
	i2c

DESCRIPTION
	Read/write I2C bus devices`
}
