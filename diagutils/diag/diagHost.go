// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package diag

import (
        "fmt"
        "time"
)

func diagHost () {

        var result, pinstate bool
        var r string

        fmt.Printf("\n%15s|%25s|%10s|%10s|%10s|%10s|%6s|%35s\n","function","parameter","units","value","min","max","result","description")
        fmt.Printf("---------------|-------------------------|----------|----------|----------|----------|------|-----------------------------------\n")

	/* diagTest: HOST_TO_BMC_INT_L interrupt 
	toggle host interrupt and validate bmc can detect the proper signal states
	*/
	//tbd: generate host interrupt 
	pinstate = gpioGet("HOST_TO_BMC_INT_L")
	if !pinstate {r = "pass"} else {r = "fail"}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","host","host_to_bmc_int_l_on","-",pinstate,active_low_on_min,active_low_on_max,r,"check interrupt is low")
	//tbd: clear interrupt
	pinstate = gpioGet("HOST_TO_BMC_INT_L")
        if pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","host","host_to_bmc_int_l_off","-",pinstate,active_low_off_min,active_low_off_max,r,"check interrupt is high")

	/* diagTest: BMC_TO_HOST_NMI
	toggle BMC to host NMI and validate host can detect the proper signal states
	*/
	gpioSet("BMC_TO_HOST_NMI",true)
        time.Sleep(50 * time.Millisecond)
	//tbd: detect pinstate at host

	gpioSet("BMC_TO_HOST_NMI",false)
        time.Sleep(50 * time.Millisecond)
        //tbd: detect pinstate at host

	/* diagTest: BMC_TO_HOST_RST_L
	Reset host and validate host behaves accordingly
	*/
	gpioSet("BMC_TO_HOST_RST_L",false)
        time.Sleep(50 * time.Millisecond)
	result = diagPing("192.168.101.131", 1)
        if !result {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","host","bmc_to_host_rst_l_on","-",result,ping_noresponse_min,ping_noresponse_max,r,"enable reset, ping host")

	gpioSet("BMC_TO_HOST_RST_L",true)
        time.Sleep(1 * time.Second)
        result = diagPing("192.168.101.128", 10)
        if result {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","host","bmc_to_host_rst_l_off","-",result,ping_response_min,ping_response_max,r,"disable reset, ping host")

	/* diagTest: BMC_TO_HOST_INT_L
	toggle BMC to host INT_L and validate host detects signal state
	*/
	gpioSet("BMC_TO_HOST_INT_L",false)
        time.Sleep(50 * time.Millisecond)
	//tbd: detect pinstate at host
	gpioSet("BMC_TO_HOST_INT_L",true)
        time.Sleep(50 * time.Millisecond)
        //tbd: detect pinstate at host

	/* diagTest: HOST_TO_BMC_I2C_GPIO  
	toggle HOST_TO_BMC_I2C_GPIO and validate bmc can detect the proper signal states
	*/

	//tbd: set HOST_TO_BMC_I2C_GPIO low
	gpioSet("CPU_TO_MAIN_I2C_EN",true)
	time.Sleep(50 * time.Millisecond)
	diagI2cWriteOffsetByte (0x00,0x74,0x02,0xF7)
	diagI2cWriteOffsetByte (0x00,0x74,0x06,0xF7)

	pinstate = gpioGet("HOST_TO_BMC_I2C_GPIO")
        if !pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","host","host_to_bmc_i2c_gpio_low","-",pinstate,active_high_off_min,active_high_off_max,r,"check gpio is high")

	//tbd: set HOST_TO_BMC_I2C_GPIO high
	diagI2cWriteOffsetByte (0x00,0x74,0x02,0xFF)
        pinstate = gpioGet("HOST_TO_BMC_I2C_GPIO")
        if pinstate {r = "pass"} else {r = "fail"}
        fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","host","host_to_bmc_i2c_gpio_high","-",pinstate,active_high_on_min,active_high_on_max,r,"check gpio is low")
	gpioSet("CPU_TO_MAIN_I2C_EN",false)

}
