// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import "github.com/platinasystems/go/internal/goes/cmd/qsfp"

func init() {
	qsfp.Init = func() {
		//port 1-16 present signals
		qsfp.VdevIo[0] = qsfp.I2cDev{0, 0x20, 0, 0x70, 0x10, 0, 0, 0}
		//port 17-32 present signals
		qsfp.VdevIo[1] = qsfp.I2cDev{0, 0x21, 0, 0x70, 0x10, 0, 0, 0}
		//port 1-16 interrupt signals
		qsfp.VdevIo[2] = qsfp.I2cDev{0, 0x22, 0, 0x70, 0x10, 0, 0, 0}
		//port 17-32 interrupt signals
		qsfp.VdevIo[3] = qsfp.I2cDev{0, 0x23, 0, 0x70, 0x10, 0, 0, 0}
		//port 1-16 LP mode signals
		qsfp.VdevIo[4] = qsfp.I2cDev{0, 0x20, 0, 0x70, 0x20, 0, 0, 0}
		//port 17-32 LP mode signals
		qsfp.VdevIo[5] = qsfp.I2cDev{0, 0x21, 0, 0x70, 0x20, 0, 0, 0}
		//port 1-16 reset signals
		qsfp.VdevIo[6] = qsfp.I2cDev{0, 0x22, 0, 0x70, 0x20, 0, 0, 0}
		//port 17-32 reset signals
		qsfp.VdevIo[7] = qsfp.I2cDev{0, 0x23, 0, 0x70, 0x20, 0, 0, 0}

		qsfp.VpageByKeyIo = map[string]uint8{
			"port-1.qsfp.presence":  0,
			"port-2.qsfp.presence":  0,
			"port-3.qsfp.presence":  0,
			"port-4.qsfp.presence":  0,
			"port-5.qsfp.presence":  0,
			"port-6.qsfp.presence":  0,
			"port-7.qsfp.presence":  0,
			"port-8.qsfp.presence":  0,
			"port-9.qsfp.presence":  0,
			"port-10.qsfp.presence": 0,
			"port-11.qsfp.presence": 0,
			"port-12.qsfp.presence": 0,
			"port-13.qsfp.presence": 0,
			"port-14.qsfp.presence": 0,
			"port-15.qsfp.presence": 0,
			"port-16.qsfp.presence": 0,
			"port-17.qsfp.presence": 1,
			"port-18.qsfp.presence": 1,
			"port-19.qsfp.presence": 1,
			"port-20.qsfp.presence": 1,
			"port-21.qsfp.presence": 1,
			"port-22.qsfp.presence": 1,
			"port-23.qsfp.presence": 1,
			"port-24.qsfp.presence": 1,
			"port-25.qsfp.presence": 1,
			"port-26.qsfp.presence": 1,
			"port-27.qsfp.presence": 1,
			"port-28.qsfp.presence": 1,
			"port-29.qsfp.presence": 1,
			"port-30.qsfp.presence": 1,
			"port-31.qsfp.presence": 1,
			"port-32.qsfp.presence": 1,
		}

		qsfp.Vdev[0] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x1}
		qsfp.Vdev[1] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x2}
		qsfp.Vdev[2] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x4}
		qsfp.Vdev[3] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x8}
		qsfp.Vdev[4] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x10}
		qsfp.Vdev[5] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x20}
		qsfp.Vdev[6] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x40}
		qsfp.Vdev[7] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x1, 0, 0x71, 0x80}
		qsfp.Vdev[8] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x1}
		qsfp.Vdev[9] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x2}
		qsfp.Vdev[10] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x4}
		qsfp.Vdev[11] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x8}
		qsfp.Vdev[12] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x10}
		qsfp.Vdev[13] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x20}
		qsfp.Vdev[14] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x40}
		qsfp.Vdev[15] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x2, 0, 0x71, 0x80}
		qsfp.Vdev[16] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x1}
		qsfp.Vdev[17] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x2}
		qsfp.Vdev[18] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x4}
		qsfp.Vdev[19] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x8}
		qsfp.Vdev[20] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x10}
		qsfp.Vdev[21] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x20}
		qsfp.Vdev[22] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x40}
		qsfp.Vdev[23] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x4, 0, 0x71, 0x80}
		qsfp.Vdev[24] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x1}
		qsfp.Vdev[25] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x2}
		qsfp.Vdev[26] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x4}
		qsfp.Vdev[27] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x8}
		qsfp.Vdev[28] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x10}
		qsfp.Vdev[29] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x20}
		qsfp.Vdev[30] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x40}
		qsfp.Vdev[31] = qsfp.I2cDev{0, 0x50, 0, 0x70, 0x8, 0, 0x71, 0x80}
	}
}
