// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package diag

import (
	"fmt"

	"github.com/platinasystems/goes/environ/nuvoton"
	"github.com/platinasystems/goes/optional/gpio"
	"github.com/platinasystems/goes/redis"
	"strconv"
	"time"
)

func diagFans() {

	var result, pinstate bool
	var r string
	var d0, d1 uint8

	const (
		w83795Bus    = 0
		w83795Adr    = 0x2f
		w83795MuxAdr = 0x76
		w83795MuxVal = 0x80
	)

	var hw = w83795.HwMonitor{w83795Bus, w83795Adr, w83795MuxAdr, w83795MuxVal}

	fmt.Printf("\n%15s|%25s|%10s|%10s|%10s|%10s|%6s|%35s\n", "function", "parameter", "units", "value", "min", "max", "result", "description")
	fmt.Printf("---------------|-------------------------|----------|----------|----------|----------|------|-----------------------------------\n")

	/* diagTest: hwm reset
	   check hwm i2c access in and out of reset
	*/
	diagI2cWrite1Byte(0x00, 0x76, 0x80)
	time.Sleep(10 * time.Millisecond)
	diagI2cWriteOffsetByte(0x00, 0x2f, 0x00, 0x00)
	diagI2cWriteOffsetByte(0x00, 0x2f, 0x01, 0x55)
	gpio.GpioSet("HWM_RST_L", false)
	time.Sleep(50 * time.Millisecond)
	result, d0 = diagI2cPing(0x00, 0x2f, 0x01, 1)
	if result && d0 == 0x80 {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "hwm_rst_l_on", "-", result, i2cping_noresponse_min, i2cping_noresponse_max, r, "enable reset, ping device")

	gpio.GpioSet("HWM_RST_L", true)
	time.Sleep(50 * time.Millisecond)
	result, d0 = diagI2cPing(0x00, 0x2f, 0x01, 1)
	if result && d0 == 0x55 {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "hwm_rst_l_off", "-", result, i2cping_response_min, i2cping_response_max, r, "disable reset, ping device")

	diagI2cWrite1Byte(0x00, 0x76, 0x00)
	hw.FanInit()
	/* diagTest: hwm interrupt
	   toggle hwm interrupt and validate bmc can detect the proper signal states
	*/
	//tbd: generate HWM interrupt
	diagI2cWrite1Byte(0x00, 0x76, 0x80)
	time.Sleep(10 * time.Millisecond)
	diagI2cWriteOffsetByte(0x00, 0x2f, 0x00, 0x00)
	diagI2cWriteOffsetByte(0x00, 0x2f, 0x40, 0x00)
	diagI2cWrite1Byte(0x00, 0x76, 0x00)
	pinstate = gpio.GpioGet("HWM_INT_L")
	if !pinstate {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "hwm_int_l_on", "-", pinstate, active_low_on_min, active_low_on_max, r, "check interrupt is low")

	gpio.GpioSet("HWM_RST_L", false)
	time.Sleep(10 * time.Millisecond)
	gpio.GpioSet("HWM_RST_L", true)
	pinstate = gpio.GpioGet("HWM_INT_L")
	if pinstate {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "hwm_int_l_off", "-", pinstate, active_low_off_min, active_low_off_max, r, "check interrupt is high")

	/* diagTest: fan board interrupt
	   Toggle interrupt and check interrupt signal state
	*/
	gpio.GpioSet("P3V3_FAN_EN", false)
	time.Sleep(50 * time.Millisecond)
	gpio.GpioSet("P3V3_FAN_EN", true)
	pinstate = gpio.GpioGet("FAN_STATUS_INT_L")
	if !pinstate {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "fan_status_int_l_off", "-", pinstate, active_low_on_min, active_low_on_max, r, "check interrupt is low")

	diagI2cWrite1Byte(0x01, 0x72, 0x04)
	time.Sleep(500 * time.Millisecond)
	result, _ = diagI2cPing(0x01, 0x20, 0x00, 1)
	result, _ = diagI2cPing(0x01, 0x20, 0x01, 1)
	diagI2cWrite1Byte(0x01, 0x72, 0x00)
	time.Sleep(10 * time.Millisecond)
	pinstate = gpio.GpioGet("FAN_STATUS_INT_L")
	if pinstate {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "fan_status_int_l_high", "-", pinstate, active_low_off_min, active_low_off_max, r, "check interrupt is high")

	/* diagTest: fan tray present
	check all fan trays are present
	*/
	diagI2cWrite1Byte(0x01, 0x72, 0x04)
	_, d0 = diagI2cPing(0x01, 0x20, 0x00, 1)
	_, d1 = diagI2cPing(0x01, 0x20, 0x01, 1)
	result = (d0&0x44 == 0) && (d1&0x44 == 0)
	if result {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "fan_tray_present", "-", result, active_high_on_min, active_high_on_max, r, "check all fan trays present")

	/* diagTest: fan tray air direction
	check all fan trays have the same air direction
	*/
	result = ((d0 & 0x88) == (d1 & 0x88)) && (((d0 & 0x88) == 0x88) || ((d0 & 0x88) == 0))
	if result {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "fan_tray_direction", "-", result, active_high_on_min, active_high_on_max, r, "check fan trays all FTB or BTF")

	/* diagTest: fan tray LEDs
	   check LED signals to each fan tray can be toggled
	*/
	diagI2cWriteOffsetByte(0x00, 0x20, 0x06, 0xcc)
	diagI2cWriteOffsetByte(0x00, 0x20, 0x07, 0xcc)
	_, d0 = diagI2cPing(0x01, 0x20, 0x00, 1)
	_, d1 = diagI2cPing(0x01, 0x20, 0x01, 1)
	result = ((d0 & 0x33) == 0x33) && ((d1 & 0x33) == 0x33)
	if result {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "fan_tray_led_high", "-", result, active_high_on_min, active_high_on_max, r, "check fan leds can drive high")

	diagI2cWriteOffsetByte(0x00, 0x20, 0x02, 0xcc)
	diagI2cWriteOffsetByte(0x00, 0x20, 0x03, 0xcc)
	_, d0 = diagI2cPing(0x01, 0x20, 0x00, 1)
	_, d1 = diagI2cPing(0x01, 0x20, 0x01, 1)
	result = ((d0 & 0x33) == 0x0) && ((d1 & 0x33) == 0x0)
	if result {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "fan_tray_led_low", "-", result, active_high_on_min, active_high_on_max, r, "check fan leds can drive low")
	diagI2cWriteOffsetByte(0x00, 0x20, 0x06, 0xff)
	diagI2cWriteOffsetByte(0x00, 0x20, 0x07, 0xff)
	diagI2cWriteOffsetByte(0x00, 0x20, 0x02, 0xff)
	diagI2cWriteOffsetByte(0x00, 0x20, 0x03, 0xff)
	diagI2cWrite1Byte(0x01, 0x72, 0x0)

	/* diagTest: fan speed
	   set fan speed and validate RPM is in expected range
	*/
	redis.Hset("platina", "fan_tray.speed", "high")
	s, err := redis.Hget("platina", "fan_tray.1.1.rpm")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	i, _ := strconv.Atoi(s)
	if i < fanspeed0_min || i > fanspeed0_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fanspeed0_rpm", "RPM", i, fanspeed0_min, fanspeed0_max, r, "set fanspeed0, check rpm")

	/* diagTest: temp sensors
	   check temp sensors are within expected range
	*/
	s, err = redis.Hget("platina", "temperature.bmc_cpu")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, _ := strconv.ParseFloat(s, 64)
	if f < tmon_bmc_cpu_min || f > tmon_bmc_cpu_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10.2f|%10.2f|%10.2f|%6s|%35s\n", "fans", "tmon_bmc_cpu", "°C", f, tmon_bmc_cpu_min, tmon_bmc_cpu_max, r, "check bmc temp sense")

	s, err = redis.Hget("platina", "temperature.fan_rear")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, _ = strconv.ParseFloat(s, 64)
	if f < tmon_fan_rear_min || f > tmon_fan_rear_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10.2f|%10.2f|%10.2f|%6s|%35s\n", "fans", "tmon_fan_rear", "°C", f, tmon_fan_rear_min, tmon_fan_rear_max, r, "check hwm exhuast temp sense")

	s, err = redis.Hget("platina", "temperature.fan_front")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, _ = strconv.ParseFloat(s, 64)
	if f < tmon_fan_front_min || f > tmon_fan_front_max {
		r = "fail"
	} else {
		r = "pass"
	}
	fmt.Printf("%15s|%25s|%10s|%10.2f|%10.2f|%10.2f|%6s|%35s\n", "fans", "tmon_fan_front", "°C", f, tmon_fan_front_min, tmon_fan_front_max, r, "check hwm intake temp sense")
}
