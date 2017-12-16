// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/cmd/platina/mk2/mc1/bmc/nct7802yd"
	"github.com/platinasystems/go/internal/gpio"
	"strconv"
)

func nct7802ydInit() {
	// Bus, Addr, MuxBus, MuxAddr, MuxValue
	nct7802yd.Vdev = nct7802yd.I2cDev{0, 0x2c, 0, 0x71, 0x04}

	cmd.Init("gpio")
	pin, found := gpio.Pins["QS_MC_SLOT_ID"]
	if found {
		r, _ := pin.Value()
		if r {
			nct7802yd.SlotId = 2
		} else {
			nct7802yd.SlotId = 1
		}
	}

	nct7802yd.VpageByKey = map[string]uint8{
		"mc." + strconv.Itoa(nct7802yd.SlotId) + ".hwmon.fan_tray.control":   0,
		"mc." + strconv.Itoa(nct7802yd.SlotId) + ".hwmon.fan_tray.speed":     0,
		"mc." + strconv.Itoa(nct7802yd.SlotId) + ".hwmon.fan_tray.duty":      0,
		"mc." + strconv.Itoa(nct7802yd.SlotId) + ".hwmon.front.temp.units.C": 0,
		"mc." + strconv.Itoa(nct7802yd.SlotId) + ".hwmon.rear.temp.units.C":  0,
		"mc." + strconv.Itoa(nct7802yd.SlotId) + ".hwmon.target.units.C":     0,
		"mc." + strconv.Itoa(nct7802yd.SlotId) + ".host.temp.units.C":        0,
		"mc." + strconv.Itoa(nct7802yd.SlotId) + ".host.temp.target.units.C": 0,
		"mc." + strconv.Itoa(nct7802yd.SlotId) + ".qsfp.temp.units.C":        0,
		"mc." + strconv.Itoa(nct7802yd.SlotId) + ".qsfp.temp.target.units.C": 0,
	}

	nct7802yd.WrRegDv["hwmon"] = "hwmon"
	nct7802yd.WrRegFn["hwmon.fan_tray.control"] = "control"
	nct7802yd.WrRegFn["hwmon.fan_tray.speed"] = "speed"
	nct7802yd.WrRegFn["hwmon.fan_tray.speed.return"] = "speed.return"
	nct7802yd.WrRegFn["hwmon.target.units.C"] = "target.units.C"

	nct7802yd.WrRegDv["host"] = "host"
	nct7802yd.WrRegFn["host.temp.units.C"] = "host.temp.units.C"
	nct7802yd.WrRegFn["host.temp.target.units.C"] = "host.temp.target.units.C"
	nct7802yd.WrRegFn["host.reset"] = "host.reset"

	nct7802yd.WrRegDv["qsfp"] = "qsfp"
	nct7802yd.WrRegFn["qsfp.temp.units.C"] = "qsfp.temp.units.C"
	nct7802yd.WrRegFn["qsfp.temp.target.units.C"] = "qsfp.temp.target.units.C"

	nct7802yd.WrRegRng["hwmon.fan_tray.control"] = []string{"enabled", "disabled"}
	nct7802yd.WrRegRng["hwmon.fan_tray.speed"] = []string{"low", "med", "high", "auto", "max"}
	nct7802yd.WrRegRng["hwmon.target.units.C"] = []string{"0", "60"}
	nct7802yd.WrRegRng["host.reset"] = []string{"true"}
}
