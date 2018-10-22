// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package diag

import (
	"fmt"
	"strconv"
	"time"

	"github.com/platinasystems/go/goes/cmd/w83795d"
	"github.com/platinasystems/redis"
)

func diagFans() error {

	var result bool
	var r string
	var d0, d1 uint8

	const (
		w83795dBus    = 0
		w83795dAdr    = 0x2f
		w83795dMuxBus = 0
		w83795dMuxAdr = 0x76
		w83795dMuxVal = 0x80
	)
	var hw = w83795d.I2cDev{w83795dBus, w83795dAdr, w83795dMuxBus, w83795dMuxAdr, w83795dMuxVal}

	fmt.Printf("\n%15s|%25s|%10s|%10s|%10s|%10s|%6s|%35s\n", "function", "parameter", "units", "value", "min", "max", "result", "description")
	fmt.Printf("---------------|-------------------------|----------|----------|----------|----------|------|-----------------------------------\n")

	// diagTest: hwm reset
	// check hwm i2c access in and out of reset
	//
	diagI2cWrite1Byte(0x00, 0x76, 0x80)
	time.Sleep(10 * time.Millisecond)
	diagI2cWriteOffsetByte(0x00, 0x2f, 0x00, 0x00)
	diagI2cWriteOffsetByte(0x00, 0x2f, 0x01, 0x55)
	gpioSet("HWM_RST_L", false)
	time.Sleep(500 * time.Millisecond)
	result, d0 = diagI2cPing(0x00, 0x2f, 0x01, 1)
	if result && d0 == 0x80 {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "hwm_rst_l_on", "-", result, i2cping_noresponse_min, i2cping_noresponse_max, r, "enable reset, ping device")

	gpioSet("HWM_RST_L", true)
	time.Sleep(500 * time.Millisecond)
	result, d0 = diagI2cPing(0x00, 0x2f, 0x01, 1)
	if result && d0 == 0x55 {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "hwm_rst_l_off", "-", result, i2cping_response_min, i2cping_response_max, r, "disable reset, ping device")

	diagI2cWrite1Byte(0x00, 0x76, 0x00)
	hw.FanInit()

	// diagTest: hwm interrupt
	// toggle hwm interrupt and validate bmc can detect the proper signal states
	//
	diagI2cWrite1Byte(0x00, 0x76, 0x80)
	time.Sleep(10 * time.Millisecond)
	diagI2cWriteOffsetByte(0x00, 0x2f, 0x00, 0x00)
	diagI2cWriteOffsetByte(0x00, 0x2f, 0x40, 0x00)
	diagI2cWrite1Byte(0x00, 0x76, 0x00)
	pinstate, _ := gpioGet("HWM_INT_L")
	r = CheckPassB(pinstate, false)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "hwm_int_l_on", "-", pinstate, active_low_on_min, active_low_on_max, r, "check interrupt is low")

	diagI2cWrite1Byte(0x00, 0x76, 0x80)
	time.Sleep(10 * time.Millisecond)
	diagI2cWriteOffsetByte(0x00, 0x2f, 0x00, 0x00)
	diagI2cWriteOffsetByte(0x00, 0x2f, 0x40, 0x10)
	pinstate, _ = gpioGet("HWM_INT_L")
	r = CheckPassB(pinstate, true)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "hwm_int_l_off", "-", pinstate, active_low_off_min, active_low_off_max, r, "check interrupt is high")

	// diagTest: fan board interrupt
	// Toggle interrupt and check interrupt signal state
	//
	gpioSet("P3V3_FAN_EN", false)
	time.Sleep(50 * time.Millisecond)
	gpioSet("P3V3_FAN_EN", true)
	pinstate, _ = gpioGet("FAN_STATUS_INT_L")
	r = CheckPassB(pinstate, false)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "fan_status_int_l_off", "-", pinstate, active_low_on_min, active_low_on_max, r, "check interrupt is low")

	time.Sleep(500 * time.Millisecond)
	diagI2cWrite1Byte(0x01, 0x72, 0x04)
	result, _ = diagI2cPing(0x01, 0x20, 0x00, 1)
	result, _ = diagI2cPing(0x01, 0x20, 0x01, 1)
	diagI2cWrite1Byte(0x01, 0x72, 0x00)
	time.Sleep(10 * time.Millisecond)
	pinstate, _ = gpioGet("FAN_STATUS_INT_L")
	r = CheckPassB(pinstate, true)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "fan_status_int_l_high", "-", pinstate, active_low_off_min, active_low_off_max, r, "check interrupt is high")

	// diagTest: fan tray present
	// check all fan trays are present
	//
	diagI2cWrite1Byte(0x01, 0x72, 0x04)
	_, d0 = diagI2cPing(0x01, 0x20, 0x00, 1)
	_, d1 = diagI2cPing(0x01, 0x20, 0x01, 1)
	result = (d0&0x44 == 0) && (d1&0x44 == 0)
	r = CheckPassB(result, true)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "fan_tray_present", "-", result, active_high_on_min, active_high_on_max, r, "check all fan trays present")

	// diagTest: fan tray air direction
	// check all fan trays have the same air direction
	//
	result = ((d0 & 0x88) == (d1 & 0x88)) && (((d0 & 0x88) == 0x88) || ((d0 & 0x88) == 0))
	r = CheckPassB(result, true)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "fan_tray_direction", "-", result, active_high_on_min, active_high_on_max, r, "check fan trays all FTB or BTF")

	// diagTest: fan tray LEDs
	// check LED signals to each fan tray can be toggled
	//
	diagI2cWriteOffsetByte(0x01, 0x20, 0x02, 0xff)
	diagI2cWriteOffsetByte(0x01, 0x20, 0x03, 0xff)
	diagI2cWriteOffsetByte(0x01, 0x20, 0x06, 0xcc)
	diagI2cWriteOffsetByte(0x01, 0x20, 0x07, 0xcc)
	_, d0 = diagI2cPing(0x01, 0x20, 0x00, 1)
	_, d1 = diagI2cPing(0x01, 0x20, 0x01, 1)
	result = ((d0 & 0x33) == 0x33) && ((d1 & 0x33) == 0x33)
	r = CheckPassB(result, true)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "fan_tray_led_high", "-", result, active_high_on_min, active_high_on_max, r, "check fan leds can drive high")

	diagI2cWriteOffsetByte(0x01, 0x20, 0x02, 0xcc)
	diagI2cWriteOffsetByte(0x01, 0x20, 0x03, 0xcc)
	_, d0 = diagI2cPing(0x01, 0x20, 0x00, 1)
	_, d1 = diagI2cPing(0x01, 0x20, 0x01, 1)
	result = ((d0 & 0x33) == 0x0) && ((d1 & 0x33) == 0x0)
	r = CheckPassB(result, true)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "fans", "fan_tray_led_low", "-", result, active_high_on_min, active_high_on_max, r, "check fan leds can drive low")
	diagI2cWriteOffsetByte(0x01, 0x20, 0x06, 0xff)
	diagI2cWriteOffsetByte(0x01, 0x20, 0x07, 0xff)
	diagI2cWriteOffsetByte(0x01, 0x20, 0x02, 0xff)
	diagI2cWriteOffsetByte(0x01, 0x20, 0x03, 0xff)
	diagI2cWrite1Byte(0x01, 0x72, 0x0)
	time.Sleep(50 * time.Millisecond)
	// diagTest: fan speed
	// set fan speed and validate RPM is in expected range
	//

	hw.SetFanSpeed("high", true)
	time.Sleep(15 * time.Second)
	p, err := hw.FanCount(1)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedhigh_min, fanspeedhigh_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan1_high_rpm", "RPM", p, fanspeedhigh_min, fanspeedhigh_max, r, "set fan speed high, check rpm")
	p, err = hw.FanCount(2)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedhigh_min, fanspeedhigh_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan2_high_rpm", "RPM", p, fanspeedhigh_min, fanspeedhigh_max, r, "set fan speed high, check rpm")
	p, err = hw.FanCount(3)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedhigh_min, fanspeedhigh_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan3_high_rpm", "RPM", p, fanspeedhigh_min, fanspeedhigh_max, r, "set fan speed high, check rpm")
	p, err = hw.FanCount(4)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedhigh_min, fanspeedhigh_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan4_high_rpm", "RPM", p, fanspeedhigh_min, fanspeedhigh_max, r, "set fan speed high, check rpm")
	p, err = hw.FanCount(5)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedhigh_min, fanspeedhigh_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan5_high_rpm", "RPM", p, fanspeedhigh_min, fanspeedhigh_max, r, "set fan speed high, check rpm")
	p, err = hw.FanCount(6)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedhigh_min, fanspeedhigh_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan6_high_rpm", "RPM", p, fanspeedhigh_min, fanspeedhigh_max, r, "set fan speed high, check rpm")
	p, err = hw.FanCount(7)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedhigh_min, fanspeedhigh_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan7_high_rpm", "RPM", p, fanspeedhigh_min, fanspeedhigh_max, r, "set fan speed high, check rpm")
	p, err = hw.FanCount(8)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedhigh_min, fanspeedhigh_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan8_high_rpm", "RPM", p, fanspeedhigh_min, fanspeedhigh_max, r, "set fan speed high, check rpm")

	hw.SetFanSpeed("med", true)
	time.Sleep(6 * time.Second)
	p, err = hw.FanCount(1)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedmed_min, fanspeedmed_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan1_med_rpm", "RPM", p, fanspeedmed_min, fanspeedmed_max, r, "set fan speed med, check rpm")
	p, err = hw.FanCount(2)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedmed_min, fanspeedmed_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan2_med_rpm", "RPM", p, fanspeedmed_min, fanspeedmed_max, r, "set fan speed med, check rpm")
	p, err = hw.FanCount(3)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedmed_min, fanspeedmed_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan3_med_rpm", "RPM", p, fanspeedmed_min, fanspeedmed_max, r, "set fan speed med, check rpm")
	p, err = hw.FanCount(4)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedmed_min, fanspeedmed_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan4_med_rpm", "RPM", p, fanspeedmed_min, fanspeedmed_max, r, "set fan speed med, check rpm")
	p, err = hw.FanCount(5)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedmed_min, fanspeedmed_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan5_med_rpm", "RPM", p, fanspeedmed_min, fanspeedmed_max, r, "set fan speed med, check rpm")
	p, err = hw.FanCount(6)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedmed_min, fanspeedmed_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan6_med_rpm", "RPM", p, fanspeedmed_min, fanspeedmed_max, r, "set fan speed med, check rpm")
	p, err = hw.FanCount(7)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedmed_min, fanspeedhigh_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan7_med_rpm", "RPM", p, fanspeedmed_min, fanspeedmed_max, r, "set fan speed med, check rpm")
	p, err = hw.FanCount(8)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedmed_min, fanspeedmed_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan8_med_rpm", "RPM", p, fanspeedmed_min, fanspeedmed_max, r, "set fan speed med, check rpm")

	hw.SetFanSpeed("low", true)
	time.Sleep(6 * time.Second)
	p, err = hw.FanCount(1)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedlow_min, fanspeedlow_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan1_low_rpm", "RPM", p, fanspeedlow_min, fanspeedlow_max, r, "set fan speed low, check rpm")
	p, err = hw.FanCount(2)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedlow_min, fanspeedlow_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan2_low_rpm", "RPM", p, fanspeedlow_min, fanspeedlow_max, r, "set fan speed low, check rpm")
	p, err = hw.FanCount(3)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedlow_min, fanspeedlow_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan3_low_rpm", "RPM", p, fanspeedlow_min, fanspeedlow_max, r, "set fan speed low, check rpm")
	p, err = hw.FanCount(4)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedlow_min, fanspeedlow_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan4_low_rpm", "RPM", p, fanspeedlow_min, fanspeedlow_max, r, "set fan speed low, check rpm")
	p, err = hw.FanCount(5)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedlow_min, fanspeedlow_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan5_low_rpm", "RPM", p, fanspeedlow_min, fanspeedlow_max, r, "set fan speed low, check rpm")
	p, err = hw.FanCount(6)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedlow_min, fanspeedlow_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan6_low_rpm", "RPM", p, fanspeedlow_min, fanspeedlow_max, r, "set fan speed low, check rpm")
	p, err = hw.FanCount(7)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedlow_min, fanspeedlow_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan7_low_rpm", "RPM", p, fanspeedlow_min, fanspeedlow_max, r, "set fan speed low, check rpm")
	p, err = hw.FanCount(8)
	if err != nil {
		return err
	}
	r = CheckPassU(p, fanspeedlow_min, fanspeedlow_max)
	fmt.Printf("%15s|%25s|%10s|%10d|%10d|%10d|%6s|%35s\n", "fans", "fan8_low_rpm", "RPM", p, fanspeedlow_min, fanspeedlow_max, r, "set fan speed low, check rpm")
	hw.SetFanSpeed("med", true)

	// diagTest: temp sensors
	// check temperature sensors are in expected range
	//
	fs, _ := redis.Hget(redis.DefaultHash, "bmc.temperature.units.C")
	f, err := strconv.ParseFloat(fs, 64)
	r = CheckPassF(f, tmon_bmc_cpu_min, tmon_bmc_cpu_max)
	fmt.Printf("%15s|%25s|%10s|%10.2f|%10.2f|%10.2f|%6s|%35s\n", "fans", "tmon_bmc_cpu", "C", f, tmon_bmc_cpu_min, tmon_bmc_cpu_max, r, "check bmc temp sense")

	v, err := hw.RearTemp()
	if err != nil {
		return err
	}
	f, _ = strconv.ParseFloat(v, 64)
	r = CheckPassF(f, tmon_fan_rear_min, tmon_fan_rear_max)
	fmt.Printf("%15s|%25s|%10s|%10.2f|%10.2f|%10.2f|%6s|%35s\n", "fans", "tmon_fan_rear", "C", f, tmon_fan_rear_min, tmon_fan_rear_max, r, "check hwm exhuast temp sense")

	v, err = hw.FrontTemp()
	if err != nil {
		return err
	}
	f, _ = strconv.ParseFloat(v, 64)
	r = CheckPassF(f, tmon_fan_front_min, tmon_fan_front_max)
	fmt.Printf("%15s|%25s|%10s|%10.2f|%10.2f|%10.2f|%6s|%35s\n", "fans", "tmon_fan_front", "C", f, tmon_fan_front_min, tmon_fan_front_max, r, "check hwm intake temp sense")

	return nil
}
