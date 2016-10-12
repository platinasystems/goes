// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package diag

import (
        "fmt"
	"time"
)

func diagFans () {

        var result, pinstate bool
	var r string
	var d uint8

	fmt.Printf("\n%15s|%25s|%10s|%10s|%10s|%10s|%6s|%35s\n","function","parameter","units","value","min","max","result","description")
        fmt.Printf("---------------|-------------------------|----------|----------|----------|----------|------|-----------------------------------\n")

        /* diagTest: hwm reset
        check hwm i2c access in and out of reset
        */
	diagI2cWrite1Byte (0x00, 0x76, 0x80)
	time.Sleep(10 * time.Millisecond)
	diagI2cWriteOffsetByte(0x00,0x2f,0x00,0x00)
	diagI2cWriteOffsetByte(0x00,0x2f,0x01,0x55)
        gpioSet("HWM_RST_L",false)
	time.Sleep(50 * time.Millisecond)
        result, d = diagI2cPing(0x00,0x2f,0x01,1)
        if result && d == 0x80 {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","fans","hwm_rst_l_on","-",result,i2cping_noresponse_min,i2cping_noresponse_max,r,"enable reset, ping device")

        gpioSet("HWM_RST_L",true)
        time.Sleep(50 * time.Millisecond)
        result, d = diagI2cPing(0x00,0x2f,0x01,1)
        if result && d == 0x55 {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","fans","hwm_rst_l_off","-",result,i2cping_response_min,i2cping_response_max,r, "disable reset, ping device")

        diagI2cWrite1Byte (0x00, 0x76, 0x00)

        /* diagTest: hwm interrupt 
        toggle hwm interrupt and validate bmc can detect the proper signal states
        */
        //tbd: generate HWM interrupt 
        pinstate = gpioGet("HWM_INT_L")
        if !pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","fans","hwm_int_l_on","-",pinstate,active_low_on_min,active_low_on_max,r,"check interrupt is low")

        //tbd" clear HWM interrupt
        pinstate = gpioGet("HWM_INT_L")
        if pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","fans","hwm_int_l_off","-",pinstate,active_low_off_min,active_low_off_max,r,"check interrupt is high")

        /* diagTest: fan board interrupt 
        check interrupt is high, cannot generate interrupt without operator action
        */
        r = "pass"
        pinstate = gpioGet("FAN_STATUS_INT_L")
	if pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","fans","fan_status_int_l_off","-",pinstate,active_low_off_min,active_low_off_max,r,"check interrupt is high")

        //tbd: toggle fan brd LEDs via i2c, use i2c
        //tbd: read fan presents via i2c, use redis
        //tbd: read fan direction via i2c, use redis
        //tbd: set/get fan speed via HWM, use redis or i2c

}

