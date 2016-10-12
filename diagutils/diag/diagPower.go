// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package diag

import (
        "fmt"
)

func diagPower () {

        var r string
	var pinstate bool
        fmt.Printf("\n%15s|%25s|%10s|%10s|%10s|%10s|%6s|%35s\n","function","parameter","units","value","min","max","result","description")
        fmt.Printf("---------------|-------------------------|----------|----------|----------|----------|------|-----------------------------------\n")

	/* diagTest: ucd reset
	check ucd i2c access in and out of reset, not used
	*//*
	gpioSet("BMC_TO_UCD_RST_L",false)
	time.Sleep(50 * time.Millisecond)
	diagI2cWrite1Byte (0x00, 0x76, 0x01)
        time.Sleep(10 * time.Millisecond)
        result = diagI2cPing(0x00,0x7e,0x00,1)
        if !result {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","power","bmc_to_ucd_rst_l_on","-",result,i2cping_noresponse_min,i2cping_noresponse_max,r,"enable reset, ping device")

	gpioSet("BMC_TO_UCD_RST_L",true)
        time.Sleep(50 * time.Millisecond)
        result = diagI2cPing(0x00,0x7e,0x00,1)
        if result {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","power","bmc_to_ucd_rst_l_off","-",result,i2cping_response_min,i2cping_response_max,r,"disable reset, ping device")

	diagI2cWrite1Byte (0x00, 0x76, 0x00)
	*/

        /* diagTest: ucd interrupt 
	toggle ucd interrupt and validate bmc can detect the proper signal states
	*/
        //tbd: generate UCD interrupt 
	pinstate = gpioGet("UCD9090_INT_L")
        if !pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","power","ucd9090_int_l_on","-",pinstate,active_low_on_min,active_low_on_max,r,"check interrupt is low")

	//tbd: clear ucd interrupt
	pinstate = gpioGet("UCD9090_INT_L")
        if pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","power","ucd9090_int_l_off","-",pinstate,active_low_on_min,active_low_on_max,r,"check interrupt is high")

        /* diagTest: ucd all_power_good
	toggle ucd all_power_good output and validate bmc can detect the proper signal states
	*/
	//tbd: set low and check ALL_PWR_GOOD is low
	pinstate = gpioGet("ALL_PWR_GOOD")
        if !pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","power","all_pwr_good_off","-",pinstate,active_high_off_min,active_high_off_max,r,"check signal is low")

        //tbd: set high and check ALL_PWR_GOOD is high
        pinstate = gpioGet("ALL_PWR_GOOD")
        if pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","power","all_pwr_good_on","-",pinstate,active_high_on_min,active_high_on_max,r,"check signal is high")

}

