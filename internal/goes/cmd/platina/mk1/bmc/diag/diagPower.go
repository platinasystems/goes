// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package diag

import (
	"fmt"
	"time"

	"github.com/platinasystems/go/internal/eeprom"
	"github.com/platinasystems/go/internal/goes/cmd/ucd9090d"
)

func diagPower() error {

	var r string

	const (
		ucd9090dBus    = 0
		ucd9090dAdr    = 0x34
		ucd9090dMuxBus = 0
		ucd9090dMuxAdr = 0x76
		ucd9090dMuxVal = 0x01
	)
	var pm = ucd9090d.I2cDev{ucd9090dBus, ucd9090dAdr, ucd9090dMuxBus, ucd9090dMuxAdr, ucd9090dMuxVal}

	d := eeprom.Device{
		BusIndex:   0,
		BusAddress: 0x55,
	}
	d.GetInfo()
	switch d.Fields.DeviceVersion {
	case 0xff:
		pm.Addr = 0x7e
	case 0x00:
		pm.Addr = 0x7e
	default:
		pm.Addr = 0x34
	}

	fmt.Printf("\n%15s|%25s|%10s|%10s|%10s|%10s|%6s|%35s\n", "function", "parameter", "units", "value", "min", "max", "result", "description")
	fmt.Printf("---------------|-------------------------|----------|----------|----------|----------|------|-----------------------------------\n")

	/*
			//tbd: generate UCD interrupt
			pinstate = gpioGet("UCD9090_INT_L")
		        if !pinstate {r = "pass"} else {r = "fail"}
		        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","power","ucd9090_int_l_on","-",pinstate,active_low_on_min,active_low_on_max,r,"check interrupt is low")

			//tbd: clear ucd interrupt
			pinstate = gpioGet("UCD9090_INT_L")
		        if pinstate {r = "pass"} else {r = "fail"}
		        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","power","ucd9090_int_l_off","-",pinstate,active_low_on_min,active_low_on_max,r,"check interrupt is high")
	*/

	/* diagTest: ucd all_power_good
	toggle ucd all_power_good output and validate bmc can detect the proper signal states
	*/
	//tbd: set high and check ALL_PWR_GOOD is high
	gpioSet("HOST_CPU_PWRBTN_L", true)
	time.Sleep(50 * time.Millisecond)
	pinstate, _ := gpioGet("ALL_PWR_GOOD")
	r = CheckPassB(pinstate, true)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "power", "all_pwr_good_on", "-", pinstate, active_high_on_min, active_high_on_max, r, "check signal is high")

	/* diagTest: ucd voltage monitoring
	check monitored voltages are within operating range
	*/
	f, err := pm.Vout(10)
	if err != nil {
		return err
	}
	r = CheckPassF(f, vmon_1v0_tha_min, vmon_1v0_tha_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v0_tha", "V", f, vmon_1v0_tha_min, vmon_1v0_tha_max, r, "check P1V0 is +/-5% nominal")

	f, err = pm.Vout(6)
	if err != nil {
		return err
	}
	r = CheckPassF(f, vmon_1v0_thc_min, vmon_1v0_thc_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v0_thc", "V", f, vmon_1v0_thc_min, vmon_1v0_thc_max, r, "check VDD_CORE is +/-5% nominal")

	f, err = pm.Vout(9)
	if err != nil {
		return err
	}
	r = CheckPassF(f, vmon_1v2_ethx_min, vmon_1v2_ethx_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v2_ethx", "V", f, vmon_1v2_ethx_min, vmon_1v2_ethx_max, r, "check P1V2 is +/-5% nominal")

	f, err = pm.Vout(8)
	if err != nil {
		return err
	}
	r = CheckPassF(f, vmon_1v25_sys_min, vmon_1v25_sys_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v25_sys", "V", f, vmon_1v25_sys_min, vmon_1v25_sys_max, r, "check P1V25 is +/-5% nominal")

	f, err = pm.Vout(7)
	if err != nil {
		return err
	}
	r = CheckPassF(f, vmon_1v8_sys_min, vmon_1v8_sys_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v8_sys", "V", f, vmon_1v8_sys_min, vmon_1v8_sys_max, r, "check P1V8 is +/-5% nominal")

	f, err = pm.Vout(4)
	if err != nil {
		return err
	}
	r = CheckPassF(f, vmon_3v3_bmc_min, vmon_3v3_bmc_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_3v3_bmc", "V", f, vmon_3v3_bmc_min, vmon_3v3_bmc_max, r, "check PERI_3V3 is +/-5% nominal")

	f, err = pm.Vout(5)
	if err != nil {
		return err
	}
	r = CheckPassF(f, vmon_3v3_sb_min, vmon_3v3_sb_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_3v3_sb", "V", f, vmon_3v3_sb_min, vmon_3v3_sb_max, r, "check P3V3_SB is +/-5% nominal")

	f, err = pm.Vout(3)
	if err != nil {
		return err
	}
	r = CheckPassF(f, vmon_3v3_sys_min, vmon_3v3_sys_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_3v3_sys", "V", f, vmon_3v3_sys_min, vmon_3v3_sys_max, r, "check P3V3 is +/-5% nominal")

	f, err = pm.Vout(2)
	if err != nil {
		return err
	}
	r = CheckPassF(f, vmon_3v8_bmc_min, vmon_3v8_bmc_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_3v8_bmc", "V", f, vmon_3v8_bmc_min, vmon_3v8_bmc_max, r, "check P3V8_BMC is +/-5% nominal")

	f, err = pm.Vout(1)
	if err != nil {
		return err
	}
	r = CheckPassF(f, vmon_5v0_sb_min, vmon_5v0_sb_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_5v0_sb", "V", f, vmon_5v0_sb_min, vmon_5v0_sb_max, r, "check P5V_SB is +/-5% nominal")

	return nil
}

func diagLoggedFaults() error {
	const (
		ucd9090dBus    = 0
		ucd9090dAdr    = 0x34
		ucd9090dMuxBus = 0
		ucd9090dMuxAdr = 0x76
		ucd9090dMuxVal = 0x01
	)
	var pm = ucd9090d.I2cDev{ucd9090dBus, ucd9090dAdr, ucd9090dMuxBus, ucd9090dMuxAdr, ucd9090dMuxVal}

	d := eeprom.Device{
		BusIndex:   0,
		BusAddress: 0x55,
	}
	d.GetInfo()
	switch d.Fields.DeviceVersion {
	case 0xff:
		pm.Addr = 0x7e
	case 0x00:
		pm.Addr = 0x7e
	default:
		pm.Addr = 0x34
	}
	log, err := pm.LoggedFaultDetail()
	if err == nil {
		fmt.Printf("%v", log)
	}
	return nil
}
