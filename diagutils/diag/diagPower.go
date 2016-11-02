// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package diag

import (
	"fmt"
	"github.com/platinasystems/go/environ/ti"
	"time"
)

func diagPower() error {

	var r string

	const (
		ucd9090Bus    = 0
		ucd9090Adr    = 0x7e
		ucd9090MuxAdr = 0x76
		ucd9090MuxVal = 0x01
	)

	var pm = ucd9090.PowerMon{ucd9090Bus, ucd9090Adr, ucd9090MuxAdr, ucd9090MuxVal}

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
	f := float64(pm.Vout(10))
	r = CheckPassF(f, vmon_1v0_tha_min, vmon_1v0_tha_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v0_tha", "V", f, vmon_1v0_tha_min, vmon_1v0_tha_max, r, "check P1V0 is +/-5% nominal")

	f = float64(pm.Vout(6))
	r = CheckPassF(f, vmon_1v0_thc_min, vmon_1v0_thc_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v0_thc", "V", f, vmon_1v0_thc_min, vmon_1v0_thc_max, r, "check VDD_CORE is +/-5% nominal")

	f = float64(pm.Vout(9))
	r = CheckPassF(f, vmon_1v2_ethx_min, vmon_1v2_ethx_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v2_ethx", "V", f, vmon_1v2_ethx_min, vmon_1v2_ethx_max, r, "check P1V2 is +/-5% nominal")

	f = float64(pm.Vout(8))
	r = CheckPassF(f, vmon_1v25_sys_min, vmon_1v25_sys_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v25_sys", "V", f, vmon_1v25_sys_min, vmon_1v25_sys_max, r, "check P1V25 is +/-5% nominal")

	f = float64(pm.Vout(7))
	r = CheckPassF(f, vmon_1v8_sys_min, vmon_1v8_sys_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v8_sys", "V", f, vmon_1v8_sys_min, vmon_1v8_sys_max, r, "check P1V8 is +/-5% nominal")

	f = float64(pm.Vout(4))
	r = CheckPassF(f, vmon_3v3_bmc_min, vmon_3v3_bmc_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_3v3_bmc", "V", f, vmon_3v3_bmc_min, vmon_3v3_bmc_max, r, "check PERI_3V3 is +/-5% nominal")

	f = float64(pm.Vout(5))
	r = CheckPassF(f, vmon_3v3_sb_min, vmon_3v3_sb_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_3v3_sb", "V", f, vmon_3v3_sb_min, vmon_3v3_sb_max, r, "check P3V3_SB is +/-5% nominal")

	f = float64(pm.Vout(3))
	r = CheckPassF(f, vmon_3v3_sys_min, vmon_3v3_sys_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_3v3_sys", "V", f, vmon_3v3_sys_min, vmon_3v3_sys_max, r, "check P3V3 is +/-5% nominal")

	f = float64(pm.Vout(2))
	r = CheckPassF(f, vmon_3v8_bmc_min, vmon_3v8_bmc_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_3v8_bmc", "V", f, vmon_3v8_bmc_min, vmon_3v8_bmc_max, r, "check P3V8_BMC is +/-5% nominal")

	f = float64(pm.Vout(1))
	r = CheckPassF(f, vmon_5v0_sb_min, vmon_5v0_sb_max)
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_5v0_sb", "V", f, vmon_5v0_sb_min, vmon_5v0_sb_max, r, "check P5V_SB is +/-5% nominal")

	return nil
}
