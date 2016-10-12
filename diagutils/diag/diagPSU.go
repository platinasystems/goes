// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package diag

import (
        "fmt"
	"time"
)

func diagPSU () {

	var pinstate bool
        var r string

        fmt.Printf("\n%15s|%25s|%10s|%10s|%10s|%10s|%6s|%35s\n","function","parameter","units","value","min","max","result","description")
        fmt.Printf("---------------|-------------------------|----------|----------|----------|----------|------|------------------------------------\n")

	/* diagTest: PSU[1:0]_PRSNT_L
	validate PSU is detected present (TBD: not present case)
	*/
	pinstate = gpioGet("PSU0_PRSNT_L")
	if !pinstate {r = "pass"} else {r = "fail"}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","psu","psu0_present_l_on","-",pinstate,active_low_on_min,active_low_on_max,r,"check psu present is low")

	pinstate = gpioGet("PSU1_PRSNT_L")
        if !pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","psu","psu1_present_l_on","-",pinstate,active_low_on_min,active_low_on_max,r,"check psu present is low")

	/* diagTest: PSU[1:0]_PWROK and PSU[1:0]_PWRON_L
    	toggle psu on and validate pwrok behaves appropriately
    	*/
	gpioSet("PSU0_PWRON_L",true)
        time.Sleep(50 * time.Millisecond)
        pinstate = gpioGet("PSU0_PWROK")
        if !pinstate {r = "pass"} else {r = "fail"}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","psu","psu0_pwron_l/pwrok_off","-",pinstate,active_high_off_min,active_high_off_max,r,"turn psu off, check psu ok is low")
	
	gpioSet("PSU0_PWRON_L",false)
        time.Sleep(50 * time.Millisecond)
        pinstate = gpioGet("PSU0_PWROK")
        if pinstate {r = "pass"} else {r = "fail"}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","psu","psu0_pwron_l/pwrok_on","-",pinstate,active_high_on_min,active_high_on_max,r,"turn psu on, check psu ok is high")

	gpioSet("PSU1_PWRON_L",true)
        time.Sleep(50 * time.Millisecond)
        pinstate = gpioGet("PSU1_PWROK")
        if !pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","psu","psu1_pwron_l/pwrok_off","-",pinstate,active_high_off_min,active_high_off_max,r,"turn psu off, check psu ok is low")

        gpioSet("PSU1_PWRON_L",false)
        time.Sleep(50 * time.Millisecond)
        pinstate = gpioGet("PSU1_PWROK")
        if pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","psu","psu1_pwron_l/pwrok_on","-",pinstate,active_high_on_min,active_high_on_max,r,"turn psu on, check psu ok is high")

	/* diagTest: PSU[1:0]_INT_L interrupt 
	toggle psu interrupt and validate bmc can detect the proper signal states
	*/
	//tbd: generate psu interrupt 
	pinstate = gpioGet("PSU0_INT_L")
        if !pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","psu","psu0_int_l_on","-",pinstate,active_low_on_min,active_low_on_max,r,"check interrupt is low")

	//tbd: clear psu interrupt
	pinstate = gpioGet("PSU0_INT_L")
        if pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","psu","psu0_int_l_off","-",pinstate,active_low_off_min,active_low_off_max,r,"check interrupt is high")

	//tbd: generate psu interrupt 
        pinstate = gpioGet("PSU1_INT_L")
        if !pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","psu","psu1_int_l_on","-",pinstate,active_low_on_min,active_low_on_max,r,"check interrupt is low")

        //tbd: clear psu interrupt
        pinstate = gpioGet("PSU1_INT_L")
        if pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","psu","psu1_int_l_off","-",pinstate,active_low_off_min,active_low_off_max,r,"check interrupt is high")
}
