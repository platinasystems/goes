// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import "github.com/platinasystems/go/internal/goes/cmd/fsp"

func init() { fsp.Init = fspInit }

func fspInit() {
	fsp.Vdev[0].Slot = 2
	fsp.Vdev[0].Bus = 1
	fsp.Vdev[0].Addr = 0x58
	fsp.Vdev[0].MuxBus = 1
	fsp.Vdev[0].MuxAddr = 0x72
	fsp.Vdev[0].MuxValue = 0x01
	fsp.Vdev[0].GpioPwrok = "PSU0_PWROK"
	fsp.Vdev[0].GpioPrsntL = "PSU0_PRSNT_L"
	fsp.Vdev[0].GpioPwronL = "PSU0_PWRON_L"
	fsp.Vdev[0].GpioIntL = "PSU0_INT_L"

	fsp.Vdev[1].Slot = 1
	fsp.Vdev[1].Bus = 1
	fsp.Vdev[1].Addr = 0x58
	fsp.Vdev[1].MuxBus = 1
	fsp.Vdev[1].MuxAddr = 0x72
	fsp.Vdev[1].MuxValue = 0x02
	fsp.Vdev[1].GpioPwrok = "PSU1_PWROK"
	fsp.Vdev[1].GpioPrsntL = "PSU1_PRSNT_L"
	fsp.Vdev[1].GpioPwronL = "PSU1_PWRON_L"
	fsp.Vdev[1].GpioIntL = "PSU1_INT_L"

	fsp.VpageByKey = map[string]uint8{

		"psu1.fan_speed.units.rpm": 1,
		"psu1.status":              1,
		"psu1.admin.state":         1,
		"psu1.mfg_id":              1,
		"psu1.mfg_model":           1,
		"psu1.i_out.units.A":       1,
		"psu1.v_in.units.V":        1,
		"psu1.v_out.units.V":       1,
		"psu1.p_out.units.W":       1,
		"psu1.p_in.units.W":        1,
		"psu1.temp1.units.C":       1,
		"psu1.temp2.units.C":       1,
		"psu2.fan_speed.units.rpm": 0,
		"psu2.status":              0,
		"psu2.admin.state":         0,
		"psu2.mfg_id":              0,
		"psu2.mfg_model":           0,
		"psu2.i_out.units.A":       0,
		"psu2.v_in.units.V":        0,
		"psu2.v_out.units.V":       0,
		"psu2.p_out.units.W":       0,
		"psu2.p_in.units.W":        0,
		"psu2.temp1.units.C":       0,
		"psu2.temp2.units.C":       0,
	}

	fsp.WrRegDv["psu1"] = "psu1"
	fsp.WrRegDv["psu2"] = "psu2"
	fsp.WrRegDv["psu"] = "psu"
	fsp.WrRegFn["psu1.example"] = "example"
	fsp.WrRegFn["psu1.admin.state"] = "admin.state"
	fsp.WrRegRng["psu1.admin.state"] = []string{"true", "false"}
	fsp.WrRegFn["psu2.example"] = "example"
	fsp.WrRegFn["psu2.admin.state"] = "admin.state"
	fsp.WrRegRng["psu2.admin.state"] = []string{"true", "false"}
	fsp.WrRegFn["psu.powercycle"] = "powercycle"
	fsp.WrRegRng["psu.powercycle"] = []string{"true"}
	fsp.WrRegRng["psu1.example"] = []string{"true", "false"}
}
