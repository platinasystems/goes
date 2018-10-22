// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	//"github.com/platinasystems/redis"
	"github.com/platinasystems/go/goes/cmd/platina/mk2/mc1/bmc/qsfpeventsd"
	"github.com/platinasystems/go/internal/gpio"
	"github.com/platinasystems/log"
)

func qsfpeventsdInit() {
	/* ****
	                var ver int
			s, err := redis.Hget(redis.DefaultHash, "eeprom.DeviceVersion")
			if err == nil {
				_, err = fmt.Sscan(s, &ver)
				if err == nil && (ver == 0 || ver == 0xff) {
					log.Print("eeprom ver: ", ver)
				}
			} else {
				log.Print("redis: ", err)
			}
			*** */

	qsfpeventsd.PortIsCopper = true
	qsfpeventsd.Old_present_n = 0x01
	qsfpeventsd.New_present_n = 0x01

	// qsfpd's io signals via pca9534 io-expander
	//      [0]=PRESENT_L, [1]=INT_L, [2]=RESET_L, [3]=LPMODE
	qsfpeventsd.VdevIo = qsfpeventsd.I2cDev{0, 0x26, 0, 0x71, 0x20, 0, 0, 0}

	// qsfpd's internal eeprom
	qsfpeventsd.VdevEp = qsfpeventsd.I2cDev{0, 0x50, 0, 0x71, 0x10, 0, 0, 0}

	gpioInit()
	pin, found := gpio.Pins["QS_MC_SLOT_ID"]
	if found {
		r, _ := pin.Value()
		if r {
			qsfpeventsd.Slotid = 2
		} else {
			qsfpeventsd.Slotid = 1
		}
		log.Print("MC slotid: ", qsfpeventsd.Slotid)
	}
}
