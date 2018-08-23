// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build linux,amd64,debug

package biosupdate

import (
	"fmt"
)

func DumpPLD() error {
	/* Dump PLD registers for debugging */
	/* Register 0x604 is 0x01 for SPI0 or 0x80 for SPI1 - Only change */
	/* Register 0x602 is 0x00 for SPI0 or 0x80 for SPI1 - reflects change to 0x604 */

	for i, addr := 0, uintptr(0x600); i < 6; i++ {
		b, err := ioReadReg8(addr + uintptr(i))
		if err != nil {
			return err
		}
		fmt.Printf("reg[%04X] = %02X\n", addr+uintptr(i), b)
	}

	return nil
}
