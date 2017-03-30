// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package diag

import (
	"fmt"
	"os"
	"time"

	"github.com/platinasystems/go/internal/goes/cmd/fantray"
	"github.com/platinasystems/go/internal/goes/cmd/platina/mk1/bmc/ledgpio"
	"github.com/platinasystems/go/internal/goes/cmd/w83795"
	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/log"
)

func diagPSU() error {

	var r string
	fmt.Printf("\n%15s|%25s|%10s|%10s|%10s|%10s|%6s|%35s\n", "function", "parameter", "units", "value", "min", "max", "result", "description")
	fmt.Printf("---------------|-------------------------|----------|----------|----------|----------|------|------------------------------------\n")

	/* diagTest: PSU[1:0]_PRSNT_L
	validate PSU is detected present (TBD: not present case)
	*/
	pinstate, _ := gpioGet("PSU0_PRSNT_L")
	r = CheckPassB(pinstate, false)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "psu", "psu0_present_l_on", "-", pinstate, active_low_on_min, active_low_on_max, r, "check psu present is low")

	pinstate, _ = gpioGet("PSU1_PRSNT_L")
	r = CheckPassB(pinstate, false)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "psu", "psu1_present_l_on", "-", pinstate, active_low_on_min, active_low_on_max, r, "check psu present is low")

	/* diagTest: PSU[1:0]_PWROK and PSU[1:0]_PWRON_L
	toggle psu on and validate pwrok behaves appropriately
	*/
	gpioSet("PSU0_PWRON_L", true)
	time.Sleep(1 * time.Second)
	pinstate, _ = gpioGet("PSU0_PWROK")
	r = CheckPassB(pinstate, false)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "psu", "psu0_pwron_l/pwrok_off", "-", pinstate, active_high_off_min, active_high_off_max, r, "turn psu off, check psu ok is low")

	gpioSet("PSU0_PWRON_L", false)
	time.Sleep(1 * time.Second)
	pinstate, _ = gpioGet("PSU0_PWROK")
	r = CheckPassB(pinstate, true)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "psu", "psu0_pwron_l/pwrok_on", "-", pinstate, active_high_on_min, active_high_on_max, r, "turn psu on, check psu ok is high")
	time.Sleep(2 * time.Second)

	gpioSet("PSU1_PWRON_L", true)
	time.Sleep(1 * time.Second)
	pinstate, _ = gpioGet("PSU1_PWROK")
	r = CheckPassB(pinstate, false)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "psu", "psu1_pwron_l/pwrok_off", "-", pinstate, active_high_off_min, active_high_off_max, r, "turn psu off, check psu ok is low")

	gpioSet("PSU1_PWRON_L", false)
	time.Sleep(1 * time.Second)
	pinstate, _ = gpioGet("PSU1_PWROK")
	r = CheckPassB(pinstate, true)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "psu", "psu1_pwron_l/pwrok_on", "-", pinstate, active_high_on_min, active_high_on_max, r, "turn psu on, check psu ok is high")

	/* diagTest: PSU[1:0]_INT_L interrupt
	Check psu interrupt is high
	*/

	pinstate, _ = gpioGet("PSU0_INT_L")
	r = CheckPassB(pinstate, true)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "psu", "psu0_int_l_off", "-", pinstate, active_low_off_min, active_low_off_max, r, "check interrupt is high")

	pinstate, _ = gpioGet("PSU1_INT_L")
	r = CheckPassB(pinstate, true)
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "psu", "psu1_int_l_off", "-", pinstate, active_low_off_min, active_low_off_max, r, "check interrupt is high")

	return nil
}

func diagPowerCycle() error {

	//i2c STOP
	sd[0] = 0
	j[0] = I{true, i2c.Write, 0, 0, sd, int(0x99), int(1), 0}
	err := DoI2cRpc()
	if err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)

	log.Print("initiate manual power cycle")
	gpioSet("PSU1_PWRON_L", true)
	gpioSet("PSU0_PWRON_L", true)
	time.Sleep(1 * time.Second)
	gpioSet("PSU1_PWRON_L", false)
	gpioSet("PSU0_PWRON_L", false)

	time.Sleep(100 * time.Millisecond)

	//i2c START
	sd[0] = 0
	j[0] = I{true, i2c.Write, 0, 0, sd, int(0x99), int(0), 0}
	err = DoI2cRpc()
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)

	log.Print("re-init fan controller")
	w83795.Vdev.Bus = 0
	w83795.Vdev.Addr = 0x2f
	w83795.Vdev.MuxBus = 0
	w83795.Vdev.MuxAddr = 0x76
	w83795.Vdev.MuxValue = 0x80
	w83795.Vdev.FanInit()

	log.Print("re-init fan trays")
	fantray.Vdev.Bus = 1
	fantray.Vdev.Addr = 0x20
	fantray.Vdev.MuxBus = 1
	fantray.Vdev.MuxAddr = 0x72
	fantray.Vdev.MuxValue = 0x04
	fantray.Vdev.FanTrayLedReinit()

	log.Print("re-init front panel LEDs")
	deviceVer, _ := readVer()
	if deviceVer == 0 || deviceVer == 1 {
		ledgpio.Vdev.Addr = 0x22
	} else {
		ledgpio.Vdev.Addr = 0x75
	}
	ledgpio.Vdev.Bus = 0
	ledgpio.Vdev.MuxBus = 0x0
	ledgpio.Vdev.MuxAddr = 0x76
	ledgpio.Vdev.MuxValue = 0x2
	ledgpio.Vdev.LedFpReinit()
	log.Print("manual power cycle complete")
	return nil
}

func readVer() (v int, err error) {
	f, err := os.Open("/tmp/ver")
	if err != nil {
		return 0, err
	}
	b1 := make([]byte, 5)
	_, err = f.Read(b1)
	if err != nil {
		return 0, err
	}
	f.Close()
	return int(b1[0]), nil
}
