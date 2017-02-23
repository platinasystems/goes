// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/internal/environ/fantray"
	"github.com/platinasystems/go/internal/environ/fsp"
	"github.com/platinasystems/go/internal/environ/nuvoton"
	"github.com/platinasystems/go/internal/environ/nxp"
	"github.com/platinasystems/go/internal/environ/ti"
)

type platform struct {
}

func (p *platform) Init() (err error) {
	p.ucd9090Init()
	p.w83795Init()
	p.fantrayInit()
	p.imx6Init()
	p.fspInit()
	if err = p.boardInit(); err != nil {
		return err
	}
	return nil
}

func (p *platform) boardInit() (err error) {
	return nil
}

func (p *platform) ucd9090Init() {
	ucd9090.Vdev.Bus = 0
	ucd9090.Vdev.Addr = 0x7e
	ucd9090.Vdev.MuxBus = 0
	ucd9090.Vdev.MuxAddr = 0x76
	ucd9090.Vdev.MuxValue = 0x01

	ucd9090.VpageByKey = map[string]uint8{
		"vmon.5v.sb":    1,
		"vmon.3v8.bmc":  2,
		"vmon.3v3.sys":  3,
		"vmon.3v3.bmc":  4,
		"vmon.3v3.sb":   5,
		"vmon.1v0.thc":  6,
		"vmon.1v8.sys":  7,
		"vmon.1v25.sys": 8,
		"vmon.1v2.ethx": 9,
		"vmon.1v0.tha":  10,
	}
}

func (p *platform) w83795Init() {
	w83795.Vdev.Bus = 0
	w83795.Vdev.Addr = 0x2f
	w83795.Vdev.MuxBus = 0
	w83795.Vdev.MuxAddr = 0x76
	w83795.Vdev.MuxValue = 0x80

	w83795.VpageByKey = map[string]uint8{
		"fan_tray.1.1.rpm": 1,
		"fan_tray.1.2.rpm": 2,
		"fan_tray.2.1.rpm": 3,
		"fan_tray.2.2.rpm": 4,
		"fan_tray.3.1.rpm": 5,
		"fan_tray.3.2.rpm": 6,
		"fan_tray.4.1.rpm": 7,
		"fan_tray.4.2.rpm": 8,
		"fan_tray.speed":   1,
	}
}

func (p *platform) fantrayInit() {
	fantray.Vdev.Bus = 1
	fantray.Vdev.Addr = 0x20
	fantray.Vdev.MuxBus = 1
	fantray.Vdev.MuxAddr = 0x72
	fantray.Vdev.MuxValue = 0x04

	fantray.VpageByKey = map[string]uint8{
		"fan_tray.1.status": 1,
		"fan_tray.2.status": 2,
		"fan_tray.3.status": 3,
		"fan_tray.4.status": 4,
	}
}

func (p *platform) imx6Init() {
	imx6.VpageByKey = map[string]uint8{
		"temperature.bmc_cpu": 1,
	}
}

func (p *platform) fspInit() {
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
		"psu1.psu_status":   0,
		"psu1.admin.state":  0,
		"psu1.mfg_id":       0,
		"psu1.mfg_model":    0,
		"psu1.page":         0,
		"psu1.status_word":  0,
		"psu1.status_vout":  0,
		"psu1.status_iout":  0,
		"psu1.status_input": 0,
		"psu1.v_in":         0,
		"psu1.i_in":         0,
		"psu1.v_out":        0,
		"psu1.i_out":        0,
		"psu1.status_temp":  0,
		"psu1.p_out":        0,
		"psu1.p_in":         0,
		"psu1.p_out_raw":    0,
		"psu1.p_in_raw":     0,
		"psu1.p_mode_raw":   0,
		"psu1.pmbus_rev":    0,
		"psu1.status_fans":  0,
		"psu1.temperature1": 0,
		"psu1.temperature2": 0,
		"psu1.fan_speed":    0,
		"psu2.psu_status":   1,
		"psu2.admin.state":  1,
		"psu2.mfg_id":       1,
		"psu2.mfg_model":    1,
		"psu2.page":         1,
		"psu2.status_word":  1,
		"psu2.status_vout":  1,
		"psu2.status_iout":  1,
		"psu2.status_input": 1,
		"psu2.v_in":         1,
		"psu2.i_in":         1,
		"psu2.v_out":        1,
		"psu2.i_out":        1,
		"psu2.status_temp":  1,
		"psu2.p_out":        1,
		"psu2.p_in":         1,
		"psu2.p_out_raw":    1,
		"psu2.p_in_raw":     1,
		"psu2.p_mode_raw":   1,
		"psu2.pmbus_rev":    1,
		"psu2.status_fans":  1,
		"psu2.temperature1": 1,
		"psu2.temperature2": 1,
		"psu2.fan_speed":    1,
	}
}
