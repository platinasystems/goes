// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	. "github.com/platinasystems/go/gpio"

	"fmt"
	"os"
)

const (
	gpio1 = 0 * 32
	gpio2 = 1 * 32
	gpio3 = 2 * 32
	gpio4 = 3 * 32
	gpio5 = 4 * 32
	gpio6 = 5 * 32
	gpio7 = 6 * 32
	gpio8 = 7 * 32
)

// Linecard gpios
var gpios = map[string]PinMap{
	"bugatti_lc": {
		"BMC_TO_MM_INT_L":       IsOutputLo | gpio1 | 8,
		"BMC_CPU_NMI":           IsOutputLo | gpio1 | 9,
		"CPU_WDT_RST_IN_L":      IsOutputLo | gpio1 | 10,
		"BMC_TO_HOST_INT_L":     IsOutputLo | gpio1 | 11,
		"LC_POWER_EN":           IsOutputLo | gpio4 | 0,
		"SET_HS_TRIP0":          IsOutputLo | gpio4 | 1,
		"SET_HS_TRIP1":          IsOutputLo | gpio4 | 2,
		"EEPROM_READY":          IsInput | gpio4 | 3,
		"PWR_GOOD_AMBER":        IsOutputLo | gpio4 | 13,
		"PWR_GOOD_BMC_UP_GREEN": IsOutputLo | gpio4 | 15,
		"I2C_SPI_BRG_INT_L":     IsInput | gpio5 | 12,
		"LC_TEM_INTK_INT_L":     IsInput | gpio5 | 13,
		"LC_TEM_EXHST_INT_L":    IsInput | gpio5 | 14,
		"LM90_INT_L":            IsInput | gpio5 | 15,
		"LC_GPIO_INT":           IsInput | gpio5 | 17,
		"LC_POWER_RST_L":        IsOutputLo | gpio6 | 6,
		"IRQ_GPIO_RST_L":        IsOutputLo | gpio6 | 7,
		"MM_I2C_MUX_RST_L":      IsOutputLo | gpio6 | 8,
		"LC_I2C_MUX_RST_L":      IsOutputLo | gpio6 | 9,
		"I2C_SPI_BRG_RST_L":     IsOutputLo | gpio6 | 10,
		"TH0_I2C_MUX_RST_L":     IsOutputLo | gpio6 | 11,
		"TH0_EXT_GPIO_INT":      IsInput | gpio6 | 18,
		"TH1_EXT_GPIO_INT":      IsInput | gpio6 | 19,
		"PMIC_INT_L":            IsInput | gpio6 | 20,
		"TH2_INT_L":             IsInput | gpio6 | 21,
		"HS_INT_L":              IsInput | gpio6 | 22,
		"TH1_I2C_MUX_RST_L":     IsOutputLo | gpio7 | 0,
		"LC_GPIO_RST":           IsOutputLo | gpio7 | 1,
		"TH0_EXT_GPIO_RST":      IsOutputLo | gpio7 | 2,
		"TH1_EXT_GPIO_RST":      IsOutputLo | gpio7 | 3,
		"5482S_RST_L":           IsOutputLo | gpio7 | 4,
		"ETHX_RST_L":            IsOutputLo | gpio7 | 5,
		"I2C1_ALERT":            IsOutputLo | gpio7 | 6,
		"PEX_FATAL_ERR_L":       IsInput | gpio7 | 7,
		"TH0_INT_L":             IsInput | gpio7 | 8,
		"TH1_INT_L":             IsInput | gpio7 | 9,
	},
	"bugatti_mm": {
		"BMC_CPU_NMI":             IsOutputLo | gpio1 | 9,
		"CPU_WDT_RST_IN_L":        IsOutputLo | gpio1 | 10,
		"BMC_TO_HOST_INT_L":       IsOutputLo | gpio1 | 11,
		"PSU0_TO_MM_PWOK_L":       IsInput | gpio1 | 12,
		"PSU1_TO_MM_PWOK_L":       IsInput | gpio1 | 13,
		"PSU2_TO_MM_PWOK_L":       IsInput | gpio1 | 14,
		"PSU3_TO_MM_PWOK_L":       IsInput | gpio1 | 15,
		"PSU4_TO_MM_PWOK_L":       IsInput | gpio1 | 16,
		"PSU5_TO_MM_PWOK_L":       IsInput | gpio1 | 17,
		"PSU6_TO_MM_PWOK_L":       IsInput | gpio1 | 18,
		"PSU7_TO_MM_PWOK_L":       IsInput | gpio1 | 19,
		"PSU8_TO_MM_PWOK_L":       IsInput | gpio1 | 20,
		"PSU9_TO_MM_PWOK_L":       IsInput | gpio1 | 21,
		"PUSHBUT_ON_OFF_LINE":     IsInput | gpio1 | 22,
		"FAN0_TO_MM_INT_L":        IsInput | gpio1 | 23,
		"FAN1_TO_MM_INT_L":        IsInput | gpio1 | 24,
		"FAN2_TO_MM_INT_L":        IsInput | gpio1 | 25,
		"MM_TO_PEER_MM_PRESENT_L": IsOutputLo | gpio2 | 0,
		"MM_TO_PEER_MM_RESET_L":   IsOutputLo | gpio2 | 1,
		"PSU0_TO_MM_PRESENT_L":    IsInput | gpio2 | 4,
		"PSU1_TO_MM_PRESENT_L":    IsInput | gpio2 | 5,
		"PSU2_TO_MM_PRESENT_L":    IsInput | gpio2 | 6,
		"PSU3_TO_MM_PRESENT_L":    IsInput | gpio2 | 7,
		"PSU4_TO_MM_PRESENT_L":    IsInput | gpio2 | 8,
		"PSU5_TO_MM_PRESENT_L":    IsInput | gpio2 | 9,
		"PSU6_TO_MM_PRESENT_L":    IsInput | gpio2 | 10,
		"PSU7_TO_MM_PRESENT_L":    IsInput | gpio2 | 11,
		"PSU8_TO_MM_PRESENT_L":    IsInput | gpio2 | 12,
		"PSU9_TO_MM_PRESENT_L":    IsInput | gpio2 | 13,
		"FAN0_TO_MM_PRESENT_L":    IsInput | gpio2 | 15,
		"FAN1_TO_MM_PRESENT_L":    IsInput | gpio2 | 16,
		"FAN2_TO_MM_PRESENT_L":    IsInput | gpio2 | 17,
		"PEER_MM_TO_MM_RESET_L":   IsInput | gpio2 | 18,
		"LTC4215_ALERT#":          IsInput | gpio3 | 0,
		"LM75_EXHAUST_OT#":        IsInput | gpio3 | 1,
		"LM75_INTAKE_OT#":         IsInput | gpio3 | 2,
		"BMC_GPIO_EXP_INT_L":      IsInput | gpio3 | 3,
		"BMC_UART_MUX1_SEL":       IsOutputLo | gpio3 | 4,
		"BMC_UART_MUX0_SEL":       IsOutputLo | gpio3 | 5,
		"BMC_UCD9090_RST_L":       IsOutputLo | gpio3 | 6,
		"BCM5482S_RST#":           IsOutputLo | gpio3 | 7,
		"SW_PCIE_RST#":            IsOutputLo | gpio3 | 8,
		"PEX_PERSTn":              IsOutputLo | gpio3 | 9,
		"BMC_MM_RST_N":            IsOutputLo | gpio3 | 10,
		"CPU_PLTRST_PCIE1_N":      IsOutputLo | gpio3 | 11,
		"CPU_PRSNT_L":             IsInput | gpio3 | 12,
		"PSU0_TO_MM_INOK_L":       IsInput | gpio3 | 16,
		"PSU1_TO_MM_INOK_L":       IsInput | gpio3 | 17,
		"PSU2_TO_MM_INOK_L":       IsInput | gpio3 | 18,
		"PSU3_TO_MM_INOK_L":       IsInput | gpio3 | 19,
		"PSU4_TO_MM_INOK_L":       IsInput | gpio3 | 20,
		"PSU5_TO_MM_INOK_L":       IsInput | gpio3 | 21,
		"PSU6_TO_MM_INOK_L":       IsInput | gpio3 | 22,
		"PSU7_TO_MM_INOK_L":       IsInput | gpio3 | 23,
		"PSU8_TO_MM_INOK_L":       IsInput | gpio3 | 24,
		"PSU9_TO_MM_INOK_L":       IsInput | gpio3 | 25,
		"FAN_SPEED_CTRL0":         IsOutputLo | gpio3 | 26,
		"FAN_SPEED_CTRL1":         IsOutputLo | gpio3 | 27,
		"FAN_SPEED_CTRL2":         IsOutputLo | gpio3 | 28,
		"MM_BI_LC0_POWER_EN":      IsInput | gpio4 | 0,
		"MM_BI_LC1_POWER_EN":      IsInput | gpio4 | 1,
		"MM_BI_LC2_POWER_EN":      IsInput | gpio4 | 2,
		"MM_BI_LC3_POWER_EN":      IsInput | gpio4 | 3,
		"MM_BI_LC4_POWER_EN":      IsInput | gpio4 | 4,
		"MM_BI_LC5_POWER_EN":      IsInput | gpio4 | 5,
		"MM_BI_LC6_POWER_EN":      IsInput | gpio4 | 6,
		"MM_BI_LC7_POWER_EN":      IsInput | gpio4 | 7,
		"MM_BI_LC8_POWER_EN":      IsInput | gpio4 | 8,
		"MM_BI_LC9_POWER_EN":      IsInput | gpio4 | 9,
		"MM_BI_LC10_POWER_EN":     IsInput | gpio4 | 10,
		"MM_BI_LC11_POWER_EN":     IsInput | gpio4 | 11,
		"MM_BI_LC12_POWER_EN":     IsInput | gpio4 | 12,
		"MM_BI_LC13_POWER_EN":     IsInput | gpio4 | 13,
		"MM_BI_LC14_POWER_EN":     IsInput | gpio4 | 14,
		"MM_BI_LC15_POWER_EN":     IsInput | gpio4 | 15,
		"MM_TO_PSU0_PSON_L":       IsOutputLo | gpio4 | 16,
		"MM_TO_PSU1_PSON_L":       IsOutputLo | gpio4 | 17,
		"MM_TO_PSU2_PSON_L":       IsOutputLo | gpio4 | 18,
		"MM_TO_PSU3_PSON_L":       IsOutputLo | gpio4 | 19,
		"MM_TO_PSU4_PSON_L":       IsOutputLo | gpio4 | 21,
		"MM_TO_PSU5_PSON_L":       IsOutputLo | gpio4 | 22,
		"MM_TO_PSU6_PSON_L":       IsOutputLo | gpio4 | 24,
		"MM_TO_PSU7_PSON_L":       IsOutputLo | gpio4 | 25,
		"MM_TO_PSU8_PSON_L":       IsOutputLo | gpio4 | 26,
		"MM_TO_PSU9_PSON_L":       IsOutputLo | gpio4 | 27,
		"MM_TO_PEER_MM_IM_ACTIVE": IsOutputLo | gpio4 | 29,
		"MM_SLOTID":               IsInput | gpio4 | 30,
		"PSU0_TO_MM_SMB_ALERT_L":  IsInput | gpio5 | 12,
		"PSU1_TO_MM_SMB_ALERT_L":  IsInput | gpio5 | 13,
		"PSU2_TO_MM_SMB_ALERT_L":  IsInput | gpio5 | 14,
		"PSU3_TO_MM_SMB_ALERT_L":  IsInput | gpio5 | 15,
		"PSU4_TO_MM_SMB_ALERT_L":  IsInput | gpio5 | 16,
		"PSU5_TO_MM_SMB_ALERT_L":  IsInput | gpio5 | 17,
		"PSU6_TO_MM_SMB_ALERT_L":  IsInput | gpio5 | 18,
		"PSU7_TO_MM_SMB_ALERT_L":  IsInput | gpio5 | 19,
		"PSU8_TO_MM_SMB_ALERT_L":  IsInput | gpio5 | 20,
		"PSU9_TO_MM_SMB_ALERT_L":  IsInput | gpio5 | 21,
		"PEER_MM_TO_MM_IM_ACTIVE": IsInput | gpio5 | 22,
		"PEER_MM_TO_MM_PRESENT_L": IsInput | gpio5 | 23,
		"LC0_TO_MM_PRESENT_L":     IsInput | gpio6 | 0,
		"LC1_TO_MM_PRESENT_L":     IsInput | gpio6 | 1,
		"LC2_TO_MM_PRESENT_L":     IsInput | gpio6 | 2,
		"LC3_TO_MM_PRESENT_L":     IsInput | gpio6 | 3,
		"LC4_TO_MM_PRESENT_L":     IsInput | gpio6 | 4,
		"LC5_TO_MM_PRESENT_L":     IsInput | gpio6 | 5,
		"LC6_TO_MM_PRESENT_L":     IsInput | gpio6 | 6,
		"LC7_TO_MM_PRESENT_L":     IsInput | gpio6 | 7,
		"LC8_TO_MM_PRESENT_L":     IsInput | gpio6 | 8,
		"LC9_TO_MM_PRESENT_L":     IsInput | gpio6 | 9,
		"LC10_TO_MM_PRESENT_L":    IsInput | gpio6 | 10,
		"LC11_TO_MM_PRESENT_L":    IsInput | gpio6 | 11,
		"LC12_TO_MM_PRESENT_L":    IsInput | gpio6 | 12,
		"LC13_TO_MM_PRESENT_L":    IsInput | gpio6 | 13,
		"LC14_TO_MM_PRESENT_L":    IsInput | gpio6 | 14,
		"LC15_TO_MM_PRESENT_L":    IsInput | gpio6 | 15,
		"LC0_TO_MM_INT_L":         IsInput | gpio6 | 16,
		"LC1_TO_MM_INT_L":         IsInput | gpio6 | 17,
		"LC2_TO_MM_INT_L":         IsInput | gpio6 | 18,
		"LC3_TO_MM_INT_L":         IsInput | gpio6 | 19,
		"LC4_TO_MM_INT_L":         IsInput | gpio6 | 20,
		"LC5_TO_MM_INT_L":         IsInput | gpio6 | 21,
		"LC6_TO_MM_INT_L":         IsInput | gpio6 | 22,
		"LC7_TO_MM_INT_L":         IsInput | gpio7 | 0,
		"LC8_TO_MM_INT_L":         IsInput | gpio7 | 1,
		"LC9_TO_MM_INT_L":         IsInput | gpio7 | 2,
		"LC10_TO_MM_INT_L":        IsInput | gpio7 | 3,
		"LC11_TO_MM_INT_L":        IsInput | gpio7 | 4,
		"LC12_TO_MM_INT_L":        IsInput | gpio7 | 5,
		"LC13_TO_MM_INT_L":        IsInput | gpio7 | 6,
		"LC14_TO_MM_INT_L":        IsInput | gpio7 | 7,
		"LC15_TO_MM_INT_L":        IsInput | gpio7 | 8,
		"MM_TO_FAN0_RESET_L":      IsOutputLo | gpio7 | 9,
		"MM_TO_FAN1_RESET_L":      IsOutputLo | gpio7 | 10,
		"MM_TO_FAN2_RESET_L":      IsOutputLo | gpio7 | 11,
	},
}

func main() {
	var err error

	SetDebugPrefix("/tmp")

	chs, err := ChipsMatching(`fsl,imx\w+-gpio`)
	if err != nil {
		panic(err)
	}
	fmt.Println(chs)
	os.Exit(0)

	pinMap := gpios["bugatti_lc"]

	pin := pinMap[os.Args[1]]
	err = pin.SetDirection()
	if err != nil {
		panic(err)
	}
	v, err := pin.Value()
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
	os.Exit(0)

	for _, pin := range pinMap {
		err = pin.SetDirection()
		if err != nil {
			panic(err)
		}
		break
	}
}
