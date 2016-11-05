// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iproc

type I2cRegs struct {
	Config Reg32

	Timing_config Reg32

	Slave_bus_address Reg32

	Fifo_control [2]Reg32

	Bit_bang_control Reg32

	_ [0x30 - 0x18]byte

	Command [2]Reg32

	Interrupt_enable Reg32

	Interrupt_status_write_1_to_clear Reg32

	Data_fifo [2]struct{ Write, Read Reg32 }
}
