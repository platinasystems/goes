// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package diag

import (
	"fmt"
	"strconv"
	//"strings"
	"time"

	"github.com/platinasystems/goes/optional/gpio"
	"github.com/platinasystems/goes/redis"
)

func diagPower() {

	var r string
	var pinstate bool
	fmt.Printf("\n%15s|%25s|%10s|%10s|%10s|%10s|%6s|%35s\n", "function", "parameter", "units", "value", "min", "max", "result", "description")
	fmt.Printf("---------------|-------------------------|----------|----------|----------|----------|------|-----------------------------------\n")
	/*
			//tbd: generate UCD interrupt
			pinstate = gpio.GpioGet("UCD9090_INT_L")
		        if !pinstate {r = "pass"} else {r = "fail"}
		        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","power","ucd9090_int_l_on","-",pinstate,active_low_on_min,active_low_on_max,r,"check interrupt is low")

			//tbd: clear ucd interrupt
			pinstate = gpio.GpioGet("UCD9090_INT_L")
		        if pinstate {r = "pass"} else {r = "fail"}
		        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","power","ucd9090_int_l_off","-",pinstate,active_low_on_min,active_low_on_max,r,"check interrupt is high")
	*/

	/* diagTest: ucd all_power_good
	toggle ucd all_power_good output and validate bmc can detect the proper signal states
	*/
	//tbd: set low and check ALL_PWR_GOOD is low
	gpio.GpioSet("HOST_CPU_PWRBTN_L", false)
	time.Sleep(50 * time.Millisecond)
	pinstate = gpio.GpioGet("ALL_PWR_GOOD")
	if !pinstate {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "power", "all_pwr_good_off", "-", pinstate, active_high_off_min, active_high_off_max, r, "check signal is low")

	//tbd: set high and check ALL_PWR_GOOD is high
	gpio.GpioSet("HOST_CPU_PWRBTN_L", true)
	time.Sleep(50 * time.Millisecond)
	pinstate = gpio.GpioGet("ALL_PWR_GOOD")
	if pinstate {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "power", "all_pwr_good_on", "-", pinstate, active_high_on_min, active_high_on_max, r, "check signal is high")

	/* diagTest: ucd voltage monitoring
	check monitored voltages are within operating range
	*/

	s, err := redis.Hget("platina", "vmon.1v0.tha")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, _ := strconv.ParseFloat(s, 64)
	if f < vmon_1v0_tha_min || f > vmon_1v0_tha_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v0_tha", "V", f, vmon_1v0_tha_min, vmon_1v0_tha_max, r, "check P1V0 is +/-5% nominal")

	s, err = redis.Hget("platina", "vmon.1v0.thc")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, _ = strconv.ParseFloat(s, 64)
	if f < vmon_1v0_thc_min || f > vmon_1v0_thc_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v0_thc", "V", f, vmon_1v0_thc_min, vmon_1v0_thc_max, r, "check VDD_CORE is +/-5% nominal")

	s, err = redis.Hget("platina", "vmon.1v2.ethx")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, _ = strconv.ParseFloat(s, 64)
	if f < vmon_1v2_ethx_min || f > vmon_1v2_ethx_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v2_ethx", "V", f, vmon_1v2_ethx_min, vmon_1v2_ethx_max, r, "check P1V2 is +/-5% nominal")

	s, err = redis.Hget("platina", "vmon.1v25.sys")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, _ = strconv.ParseFloat(s, 64)
	if f < vmon_1v25_sys_min || f > vmon_1v25_sys_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v25_sys", "V", f, vmon_1v25_sys_min, vmon_1v25_sys_max, r, "check P1V25 is +/-5% nominal")

	s, err = redis.Hget("platina", "vmon.1v8.sys")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, _ = strconv.ParseFloat(s, 64)
	if f < vmon_1v8_sys_min || f > vmon_1v8_sys_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_1v8_sys", "V", f, vmon_1v8_sys_min, vmon_1v8_sys_max, r, "check P1V8 is +/-5% nominal")

	s, err = redis.Hget("platina", "vmon.3v3.bmc")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, _ = strconv.ParseFloat(s, 64)
	if f < vmon_3v3_bmc_min || f > vmon_3v3_bmc_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_3v3_bmc", "V", f, vmon_3v3_bmc_min, vmon_3v3_bmc_max, r, "check PERI_3V3 is +/-5% nominal")

	s, err = redis.Hget("platina", "vmon.3v3.sb")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, _ = strconv.ParseFloat(s, 64)
	if f < vmon_3v3_sb_min || f > vmon_3v3_sb_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_3v3_sb", "V", f, vmon_3v3_sb_min, vmon_3v3_sb_max, r, "check P3V3_SB is +/-5% nominal")

	s, err = redis.Hget("platina", "vmon.3v3.sys")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, _ = strconv.ParseFloat(s, 64)
	if f < vmon_3v3_sys_min || f > vmon_3v3_sys_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_3v3_sys", "V", f, vmon_3v3_sys_min, vmon_3v3_sys_max, r, "check P3V3 is +/-5% nominal")

	s, err = redis.Hget("platina", "vmon.3v8.bmc")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, _ = strconv.ParseFloat(s, 64)
	if f < vmon_3v8_bmc_min || f > vmon_3v8_bmc_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_3v8_bmc", "V", f, vmon_3v8_bmc_min, vmon_3v8_bmc_max, r, "check P3V8_BMC is +/-5% nominal")

	s, err = redis.Hget("platina", "vmon.5v.sb")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, _ = strconv.ParseFloat(s, 64)
	if f < vmon_5v0_sb_min || f > vmon_5v0_sb_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10.3f|%10.3f|%10.3f|%6s|%35s\n", "power", "vmon_5v0_sb", "V", f, vmon_5v0_sb_min, vmon_5v0_sb_max, r, "check P5V_SB is +/-5% nominal")

}
