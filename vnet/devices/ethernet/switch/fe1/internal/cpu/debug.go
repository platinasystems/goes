// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build debug

package cpu

import (
	. "github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/debug"
	"unsafe"
)

var basePointer = unsafe.Pointer(&struct{}{})

func check(tag string, p unsafe.Pointer, expect uint) {
	CheckRegAddr(tag, uint(uintptr(p)-uintptr(basePointer)), expect)
}

// Check memory map.
func init() {
	r := (*regs)(basePointer)
	check("miim", unsafe.Pointer(&r.miim), 0x11000)
	check("rx_buf", unsafe.Pointer(&r.rx_buf), 0x1a000)
	check("tx_buf", unsafe.Pointer(&r.tx_buf), 0x1b000)
	check("led0", unsafe.Pointer(&r.led0), 0x20000)
	check("led1", unsafe.Pointer(&r.led1), 0x21000)
	check("sub_controllers[0]", unsafe.Pointer(&r.sub_controllers[0]), 0x31000)
	check("packet dma", unsafe.Pointer(&r.sub_controllers[0].packet_dma), 0x31110)
}
