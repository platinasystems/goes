// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/goes/cmd/platina/mk2/mc1/bmc/w83795d"
)

func init() { w83795d.Init = w83795dInit }

func w83795dInit() {

	// Bus, Addr, MuxBus, MuxAddr, MuxValue, MuxBus2, MuxAddr2, MuxValue2
	w83795d.Vdev[0] = w83795d.I2cDev{1, 0x2f, 1, 0x70, 0x08, 1, 0x73, 0x10}
	w83795d.Vdev[1] = w83795d.I2cDev{1, 0x2f, 1, 0x70, 0x08, 1, 0x73, 0x20}
	w83795d.Vdev[2] = w83795d.I2cDev{1, 0x2f, 1, 0x70, 0x08, 1, 0x73, 0x40}
	w83795d.Vdev[3] = w83795d.I2cDev{1, 0x2f, 1, 0x70, 0x08, 1, 0x73, 0x80}
	w83795d.Vdev[4] = w83795d.I2cDev{1, 0x2f, 1, 0x70, 0x10, 1, 0x73, 0x04}
	w83795d.Vdev[5] = w83795d.I2cDev{1, 0x2f, 1, 0x70, 0x10, 1, 0x73, 0x08}

	// Bus, Addr, MuxBus, MuxAddr, MuxValue, MuxBus2, MuxAddr2, MuxValue2
	w83795d.VdevIo[0] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x08, 1, 0x73, 0x10}
	w83795d.VdevIo[1] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x08, 1, 0x73, 0x20}
	w83795d.VdevIo[2] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x08, 1, 0x73, 0x40}
	w83795d.VdevIo[3] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x08, 1, 0x73, 0x80}
	w83795d.VdevIo[4] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x10, 1, 0x73, 0x04}
	w83795d.VdevIo[5] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x10, 1, 0x73, 0x08}

	w83795d.WrRegDv["fan_tray.1"] = "fan_tray.1"
	w83795d.WrRegFn["fan_tray.1.example"] = "example"
	w83795d.WrRegFn["fan_tray.1.speed"] = "speed"
	w83795d.WrRegFn["fan_tray.1.control"] = "control"
	w83795d.WrRegFn["fan_tray.1.speed.return"] = "speed.return"
	w83795d.WrRegFn["fan_tray.1.hwmon.target.units.C"] = "target.units.C"

	w83795d.WrRegDv["fan_tray.2"] = "fan_tray.2"
	w83795d.WrRegFn["fan_tray.2.example"] = "example"
	w83795d.WrRegFn["fan_tray.2.speed"] = "speed"
	w83795d.WrRegFn["fan_tray.2.control"] = "control"
	w83795d.WrRegFn["fan_tray.2.speed.return"] = "speed.return"
	w83795d.WrRegFn["fan_tray.2.hwmon.target.units.C"] = "target.units.C"

	w83795d.WrRegDv["fan_tray.3"] = "fan_tray.3"
	w83795d.WrRegFn["fan_tray.3.example"] = "example"
	w83795d.WrRegFn["fan_tray.3.speed"] = "speed"
	w83795d.WrRegFn["fan_tray.3.control"] = "control"
	w83795d.WrRegFn["fan_tray.3.speed.return"] = "speed.return"
	w83795d.WrRegFn["fan_tray.3.hwmon.target.units.C"] = "target.units.C"

	w83795d.WrRegDv["fan_tray.4"] = "fan_tray.4"
	w83795d.WrRegFn["fan_tray.4.example"] = "example"
	w83795d.WrRegFn["fan_tray.4.speed"] = "speed"
	w83795d.WrRegFn["fan_tray.4.control"] = "control"
	w83795d.WrRegFn["fan_tray.4.speed.return"] = "speed.return"
	w83795d.WrRegFn["fan_tray.4.hwmon.target.units.C"] = "target.units.C"

	w83795d.WrRegDv["fan_tray.5"] = "fan_tray.5"
	w83795d.WrRegFn["fan_tray.5.example"] = "example"
	w83795d.WrRegFn["fan_tray.5.speed"] = "speed"
	w83795d.WrRegFn["fan_tray.5.control"] = "control"
	w83795d.WrRegFn["fan_tray.5.speed.return"] = "speed.return"
	w83795d.WrRegFn["fan_tray.5.hwmon.target.units.C"] = "target.units.C"

	w83795d.WrRegDv["fan_tray.6"] = "fan_tray.6"
	w83795d.WrRegFn["fan_tray.6.example"] = "example"
	w83795d.WrRegFn["fan_tray.6.speed"] = "speed"
	w83795d.WrRegFn["fan_tray.6.control"] = "control"
	w83795d.WrRegFn["fan_tray.6.speed.return"] = "speed.return"
	w83795d.WrRegFn["fan_tray.6.hwmon.target.units.C"] = "target.units.C"

	w83795d.WrRegRng["fan_tray.1.speed"] = []string{"low", "med", "high", "auto", "max"}
	w83795d.WrRegRng["fan_tray.1.control"] = []string{"local", "remote.mc1", "remote.mc2"}
	w83795d.WrRegRng["fan_tray.1.hwmon.target.units.C"] = []string{"0", "60"}

	w83795d.WrRegRng["fan_tray.2.speed"] = []string{"low", "med", "high", "auto", "max"}
	w83795d.WrRegRng["fan_tray.2.control"] = []string{"local", "remote.mc1", "remote.mc2"}
	w83795d.WrRegRng["fan_tray.2.hwmon.target.units.C"] = []string{"0", "60"}

	w83795d.WrRegRng["fan_tray.3.speed"] = []string{"low", "med", "high", "auto", "max"}
	w83795d.WrRegRng["fan_tray.3.control"] = []string{"local", "remote.mc1", "remote.mc2"}
	w83795d.WrRegRng["fan_tray.3.hwmon.target.units.C"] = []string{"0", "60"}

	w83795d.WrRegRng["fan_tray.4.speed"] = []string{"low", "med", "high", "auto", "max"}
	w83795d.WrRegRng["fan_tray.4.control"] = []string{"local", "remote.mc1", "remote.mc2"}
	w83795d.WrRegRng["fan_tray.4.hwmon.target.units.C"] = []string{"0", "60"}

	w83795d.WrRegRng["fan_tray.5.speed"] = []string{"low", "med", "high", "auto", "max"}
	w83795d.WrRegRng["fan_tray.5.control"] = []string{"local", "remote.mc1", "remote.mc2"}
	w83795d.WrRegRng["fan_tray.5.hwmon.target.units.C"] = []string{"0", "60"}

	w83795d.WrRegRng["fan_tray.6.speed"] = []string{"low", "med", "high", "auto", "max"}
	w83795d.WrRegRng["fan_tray.6.control"] = []string{"local", "remote.mc1", "remote.mc2"}
	w83795d.WrRegRng["fan_tray.6.hwmon.target.units.C"] = []string{"0", "60"}
}
