package diag

import (
	"fmt"
	"time"
)

func diagNetwork() error {

	var r string
	var result bool

	fmt.Printf("\n%15s|%25s|%10s|%10s|%10s|%10s|%6s|%35s\n", "function", "parameter", "units", "value", "min", "max", "result", "description")
	fmt.Printf("---------------|-------------------------|----------|----------|----------|----------|------|-----------------------------------\n")

	/* diagTest: ETHX_INT_L interrupt: toggle ethx interrupt and validate bmc can detect the proper signal states */
	//tbd: generate ethx interrupt
	//tbd: use cat /proc/interrupts instead?
	pinstate, err := gpioGet("ETHX_INT_L")
	if err != nil {
		return err
	}
	if !pinstate {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "mgmt_network", "ethx_int_l_on", "-", pinstate, active_low_on_min, active_low_on_max, r, "check interrupt is low")

	//tbd: clear ethx interrupt
	pinstate, err = gpioGet("ETHX_INT_L")
	if err != nil {
		return err
	}
	if pinstate {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "mgmt_network", "ethx_int_l_off", "-", pinstate, active_low_off_min, active_low_off_max, r, "check interrupt is high")

	/* diagTest: ETHX_RST_L: toggle BMC to ethx RST_L and validate ethx behaves accordingly */
	/*
	   gpioSet("ETHX_RST_L",false)
	   time.Sleep(50 * time.Millisecond)
	   result = diagPing("192.168.101.1", 1)
	   if !result {r = "pass"} else {r = "fail"}
	   fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n","network","ethx_rst_l_on","-",result,ping_noresponse,ping_noresponse_max,r,"enable reset, ping external host")
	*/
	gpioSet("ETHX_RST_L", true)
	time.Sleep(50 * time.Millisecond)
	result = diagPing("192.168.101.1", 10)
	if result {
		r = "pass"
	} else {
		r = "fail"
	}
	fmt.Printf("%15s|%25s|%10s|%10t|%10t|%10t|%6s|%35s\n", "mgmt_network", "ethx_rst_l_off", "-", result, ping_response_min, ping_response_max, r, "disable reset, ping external host")

	/* diagTest: ethx MDIO, tbd: validate reads/writes across MDIO interface */

	return nil
}
