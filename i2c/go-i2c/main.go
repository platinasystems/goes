// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package main

import (
	"fmt"

	"github.com/platinasystems/go/i2c"
)

func main() {
	var d i2c.SMBusData
	b := &i2c.Bus{}
	err := b.Open(2)
	if err != nil {
		panic(err)
	}
	err = b.SetSlaveAddress(0x50)
	if err != nil {
		panic(err)
	}
	err = b.Read(10, i2c.ByteData, &d)
	if err != nil {
		panic(err)
	}
	fmt.Println(d[:])
}
