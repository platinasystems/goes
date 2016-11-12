// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icpu

type I2cRegs struct {
	Config U32

	Timing_config U32

	Slave_bus_address U32

	Fifo_control [2]U32

	Bit_bang_control U32

	_ [0x30 - 0x18]byte

	Command [2]U32

	Interrupt_enable U32

	Interrupt_status_write_1_to_clear U32

	Data_fifo [2]struct{ Write, Read U32 }
}
