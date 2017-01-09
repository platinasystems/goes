// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build arm

// This is an example Baseboard Management Controller.
package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/platinasystems/go/internal/eeprom"
	"github.com/platinasystems/go/internal/fdt"
	"github.com/platinasystems/go/internal/fdtgpio"
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/gpio"
	optgpio "github.com/platinasystems/go/internal/optional/gpio"
	"github.com/platinasystems/go/internal/optional/i2c"
	//"github.com/platinasystems/go/internal/optional/platina-mk1/toggle"
	"github.com/platinasystems/go/internal/optional/telnetd"
	"github.com/platinasystems/go/internal/optional/watchdog"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/diag"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ/fantray"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ/fsp"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ/nuvoton"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ/nxp"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ/ti"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/info"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/led"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/machined"
	"github.com/platinasystems/go/internal/required"
	"github.com/platinasystems/go/internal/required/nld"
	"github.com/platinasystems/go/internal/required/redisd"
)

type parser interface {
	Parse(string) error
}

type Info struct {
	done     chan<- struct{}
	name     string
	prefixes []string
	attrs    machined.Attrs
}

type funcS func(string)
type funcI func(int)
type funcU16 func(uint16)
type funcF64 func(float64)

const (
	AddressBytes  = 6
	w83795Bus     = 0
	w83795Adr     = 0x2f
	w83795MuxBus  = 0
	w83795MuxAdr  = 0x76
	w83795MuxVal  = 0x80
	ucd9090Bus    = 0
	ucd9090Adr    = 0x7e
	ucd9090MuxBus = 0
	ucd9090MuxAdr = 0x76
	ucd9090MuxVal = 0x01
	ledgpioBus    = 0
	ledgpioAdr    = 0x22
	ledgpioMuxBus = 0
	ledgpioMuxAdr = 0x76
	ledgpioMuxVal = 0x02
	ledgpioReset  = "SYSTEM_LED_RST_L"
	fangpioBus    = 1
	fangpioAdr    = 0x20
	fangpioMuxBus = 1
	fangpioMuxAdr = 0x72
	fangpioMuxVal = 0x04
	ps1Slot       = 2
	ps1Bus        = 1
	ps1Adr        = 0x58
	ps1MuxBus     = 1
	ps1MuxAdr     = 0x72
	ps1MuxVal     = 0x01
	ps1GpioPwrok  = "PSU0_PWROK"
	ps1GpioPrsntL = "PSU0_PRSNT_L"
	ps1GpioPwronL = "PSU0_PWRON_L"
	ps1GpioIntL   = "PSU0_INT_L"
	ps2Slot       = 1
	ps2Bus        = 1
	ps2Adr        = 0x58
	ps2MuxBus     = 1
	ps2MuxAdr     = 0x72
	ps2MuxVal     = 0x02
	ps2GpioPwrok  = "PSU1_PWROK"
	ps2GpioPrsntL = "PSU1_PRSNT_L"
	ps2GpioPwronL = "PSU1_PWRON_L"
	ps2GpioIntL   = "PSU1_INT_L"

	redis_default_interval = 5
)

type Address [AddressBytes]byte

var timerMutex = &sync.Mutex{}
var tckr *time.Ticker
var redis_sync uint16
var redis_interval uint16 = redis_default_interval
var redis_clear string

var hw = w83795.HwMonitor{w83795Bus, w83795Adr, w83795MuxBus, w83795MuxAdr, w83795MuxVal}
var pm = ucd9090.PMon{ucd9090Bus, ucd9090Adr, ucd9090MuxBus, ucd9090MuxAdr, ucd9090MuxVal}
var ledfp = led.LedCon{ledgpioBus, ledgpioAdr, ledgpioMuxBus, ledgpioMuxAdr, ledgpioMuxVal, ledgpioReset}
var fanTray = fantray.FanStat{fangpioBus, fangpioAdr, fangpioMuxBus, fangpioMuxAdr, fangpioMuxVal}
var ps2 = fsp.Psu{ps1Slot, 0, " ", ps1Bus, ps1Adr, ps1MuxBus, ps1MuxAdr, ps1MuxVal, ps1GpioPwrok, ps1GpioPrsntL, ps1GpioPwronL, ps1GpioIntL}
var ps1 = fsp.Psu{ps2Slot, 0, " ", ps2Bus, ps2Adr, ps2MuxBus, ps2MuxAdr, ps2MuxVal, ps2GpioPwrok, ps2GpioPrsntL, ps2GpioPwronL, ps2GpioIntL}
var cpu = imx6.Cpu{}

var RedisEnvShadow = map[string]interface{}{}
var regWriteString map[string]func(string)
var regWriteInt map[string]func(int)
var regWriteUint16 map[string]func(uint16)
var regWriteFloat64 map[string]func(float64)

var stageString string
var stageKeyString string
var stageFlagString int = 0
var stageInt int
var stageKeyInt string
var stageFlagInt int = 0
var stageUint16 uint16
var stageKeyUint16 string
var stageFlagUint16 int = 0
var stageFloat64 float64
var stageKeyFloat64 string
var stageFlagFloat64 int = 0

func main() {
	gpio.File = "/boot/platina-mk1-bmc.dtb"
	g := make(goes.ByName)
	g.Plot(required.New()...)
	g.Plot(
		diag.New(),
		optgpio.New(),
		i2c.New(),
		// FIXME: remove machined after converting remaining info
		machined.New(),
		telnetd.New(),
		//toggle.New(),
		watchdog.New(),
	)
	redisd.Machine = "platina-mk1-bmc"
	redisd.Devs = []string{"lo", "eth0"}
	nld.Prefixes = []string{"lo.", "eth0."}
	machined.Hook = hook
	if err := g.Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func enableToggle() error {
	time.Sleep(time.Second * time.Duration(45))

	gpio.Aliases = make(gpio.GpioAliasMap)
	gpio.Pins = make(gpio.PinMap)

	// Parse linux.dtb to generate gpio map for this machine
	if b, err := ioutil.ReadFile(gpio.File); err == nil {
		t := &fdt.Tree{Debug: false, IsLittleEndian: false}
		t.Parse(b)

		t.MatchNode("aliases", fdtgpio.GatherAliases)
		t.EachProperty("gpio-controller", "", fdtgpio.GatherPins)
	} else {
		return fmt.Errorf("%s: %v", gpio.File, err)
	}

	pin, found := gpio.Pins["CPU_TO_MAIN_I2C_EN"]
	if !found {
		return fmt.Errorf("%s: not found", "CPU_TO_MAIN_I2C_EN")
	}
	pin.SetValue(true)
	i2c.WriteByte(0, 0x74, 2, 0xff)
	i2c.WriteByte(0, 0x74, 6, 0xdf)
	pin, found = gpio.Pins["FP_BTN_UARTSEL_EN_L"]
	if !found {
		return fmt.Errorf("%s: not found", "FP_BTN_UARTSEL_EN_L")
	}
	pin.SetValue(true)
	return nil
}

func hook() error {
	regWriteString = make(map[string]func(string))
	regWriteInt = make(map[string]func(int))
	regWriteUint16 = make(map[string]func(uint16))
	regWriteFloat64 = make(map[string]func(float64))

	// Set HW register write regWrite function
	regWriteUint16["psu1.page"] = funcU16(ps1.PageWr)
	regWriteUint16["psu2.page"] = funcU16(ps2.PageWr)
	regWriteString["fan_tray.speed"] = funcS(hw.SetFanSpeed)
	regWriteString["psu1.admin.state"] = funcS(ps1.SetAdminState)
	regWriteString["psu2.admin.state"] = funcS(ps2.SetAdminState)
	regWriteUint16["update.sync"] = funcU16(force_sync)
	regWriteUint16["update.interval"] = funcU16(change_interval)
	regWriteString["update.counters"] = funcS(clear_counters)

	gpio.Aliases = make(gpio.GpioAliasMap)
	gpio.Pins = make(gpio.PinMap)

	// Parse linux.dtb to generate gpio map for this machine
	if b, err := ioutil.ReadFile(gpio.File); err == nil {
		t := &fdt.Tree{Debug: false, IsLittleEndian: false}
		t.Parse(b)

		t.MatchNode("aliases", fdtgpio.GatherAliases)
		t.EachProperty("gpio-controller", "", fdtgpio.GatherPins)
	} else {
		return fmt.Errorf("%s: %v", gpio.File, err)
	}

	// Set gpio input/output as defined in dtb
	for name, pin := range gpio.Pins {
		err := pin.SetDirection()
		if err != nil {
			fmt.Printf("%s: %v\n", name, err)
		}
	}

	d := eeprom.Device{
		BusIndex:   0,
		BusAddress: 0x55,
	}
	var dd Address
	if e := d.GetInfo(); e != nil {
		t := time.Now()
		n := t.Nanosecond()
		r := rand.New(rand.NewSource(int64(n)))
		dd[0] = uint8(0x02)
		dd[1] = uint8(0x46)
		dd[2] = uint8(0x8a)
		dd[3] = uint8(r.Uint32() & 0xff)
		dd[4] = uint8(r.Uint32() & 0xff)
		dd[5] = uint8(r.Uint32() & 0xff)
	} else {
		dd = d.Fields.BaseEthernetAddress
		g := (uint32(dd[3])<<16 | uint32(dd[4])<<8 | uint32(dd[5])) + uint32(d.Fields.NEthernetAddress)
		dd[3] = uint8((g & 0xff0000) >> 16)
		dd[4] = uint8((g & 0xff00) >> 8)
		dd[5] = uint8(g & 0xff)
	}

	switch d.Fields.DeviceVersion {
	case 0xff:
		pm.Addr = 0x7e
		ledfp.Addr = 0x22
	case 0x00:
		pm.Addr = 0x7e
		ledfp.Addr = 0x22
	default:
		pm.Addr = 0x34
		ledfp.Addr = 0x75
	}

	hw.FanInit()
	ledfp.LedFpInit()
	fanTray.FanTrayLedInit()

	//go enableToggle()

	machined.Plot(
		&Info{
			name:     "fan",
			prefixes: []string{"fan."},
			attrs: machined.Attrs{
				"fan.front": 100,
				"fan.rear":  100,
			},
		},
		&Info{
			name:     "mfg",
			prefixes: []string{"mfg."},
			attrs: machined.Attrs{
				"mfg.platform.name":           d.Fields.PlatformName,
				"mfg.vendor":                  d.Fields.Vendor,
				"mfg.part.number":             d.Fields.PartNumber,
				"mfg.serial.number":           d.Fields.SerialNumber,
				"mfg.device.version":          d.Fields.DeviceVersion,
				"mfg.manufacture.date":        d.Fields.ManufactureDate,
				"mfg.main.pcba.part.number":   d.Fields.PcbaPartNumber,
				"mfg.main.pcba.serial.number": d.Fields.Tor1MainPcbaSerialNumber,
				"mfg.cpu.pcba.serial.number":  d.Fields.Tor1CpuPcbaSerialNumber,
				"mfg.fan.pcba.serial.number":  d.Fields.Tor1FanPcbaSerialNumber,
			},
		},
		&Info{
			name:     "update",
			prefixes: []string{"update."},
			attrs: machined.Attrs{
				"update.sync":     init_sync(),
				"update.interval": init_interval(),
				"update.counters": init_counters(),
			},
		},
		&Info{
			name:     "vmon",
			prefixes: []string{"vmon."},
			attrs: machined.Attrs{
				"vmon.5v.sb":    pm.Vout(1),
				"vmon.3v8.bmc":  pm.Vout(2),
				"vmon.3v3.sys":  pm.Vout(3),
				"vmon.3v3.bmc":  pm.Vout(4),
				"vmon.3v3.sb":   pm.Vout(5),
				"vmon.1v0.thc":  pm.Vout(6),
				"vmon.1v8.sys":  pm.Vout(7),
				"vmon.1v25.sys": pm.Vout(8),
				"vmon.1v2.ethx": pm.Vout(9),
				"vmon.1v0.tha":  pm.Vout(10),
			},
		},
		&Info{
			name:     "chassis",
			prefixes: []string{"fan_tray."},
			attrs: machined.Attrs{
				"fan_tray.1.1.rpm":  hw.FanCount(1),
				"fan_tray.1.2.rpm":  hw.FanCount(2),
				"fan_tray.2.1.rpm":  hw.FanCount(3),
				"fan_tray.2.2.rpm":  hw.FanCount(4),
				"fan_tray.3.1.rpm":  hw.FanCount(5),
				"fan_tray.3.2.rpm":  hw.FanCount(6),
				"fan_tray.4.1.rpm":  hw.FanCount(7),
				"fan_tray.4.2.rpm":  hw.FanCount(8),
				"fan_tray.1.status": fanTray.FanTrayStatus(1),
				"fan_tray.2.status": fanTray.FanTrayStatus(2),
				"fan_tray.3.status": fanTray.FanTrayStatus(3),
				"fan_tray.4.status": fanTray.FanTrayStatus(4),
				"fan_tray.speed":    hw.GetFanSpeed(),
			},
		},
		&Info{
			name:     "psu1",
			prefixes: []string{"psu1."},
			attrs: machined.Attrs{
				"psu1.status":       ps1.PsuStatus(),
				"psu1.admin.state":  ps1.GetAdminState(),
				"psu1.mfg_id":       ps1.MfgIdent(),
				"psu1.mfg_model":    ps1.MfgModel(),
				"psu1.page":         uint16(ps1.Page()),
				"psu1.status_word":  ps1.StatusWord(),
				"psu1.status_vout":  ps1.StatusVout(),
				"psu1.status_iout":  ps1.StatusIout(),
				"psu1.status_input": ps1.StatusInput(),
				"psu1.v_in":         ps1.Vin(),
				"psu1.i_in":         ps1.Iin(),
				"psu1.v_out":        ps1.Vout(),
				"psu1.i_out":        ps1.Iout(),
				"psu1.status_temp":  ps1.StatusTemp(),
				"psu1.p_out":        ps1.Pout(),
				"psu1.p_in":         ps1.Pin(),
				"psu1.p_out_raw":    ps1.PoutRaw(),
				"psu1.p_in_raw":     ps1.PinRaw(),
				"psu1.p_mode_raw":   ps1.ModeRaw(),
				"psu1.pmbus_rev":    ps1.PMBusRev(),
				"psu1.status_fans":  ps1.StatusFans(),
				"psu1.temperature1": ps1.Temp1(),
				"psu1.temperature2": ps1.Temp2(),
				"psu1.fan_speed":    ps1.FanSpeed(),
			},
		},
		&Info{
			name:     "psu2",
			prefixes: []string{"psu2."},
			attrs: machined.Attrs{
				"psu2.status":       ps2.PsuStatus(),
				"psu2.admin.state":  ps2.GetAdminState(),
				"psu2.mfg_id":       ps2.MfgIdent(),
				"psu2.mfg_model":    ps2.MfgModel(),
				"psu2.page":         uint16(ps2.Page()),
				"psu2.status_word":  ps2.StatusWord(),
				"psu2.status_vout":  ps2.StatusVout(),
				"psu2.status_iout":  ps2.StatusIout(),
				"psu2.status_input": ps2.StatusInput(),
				"psu2.v_in":         ps2.Vin(),
				"psu2.i_in":         ps2.Iin(),
				"psu2.v_out":        ps2.Vout(),
				"psu2.i_out":        ps2.Iout(),
				"psu2.status_temp":  ps2.StatusTemp(),
				"psu2.p_out":        ps2.Pout(),
				"psu2.p_in":         ps2.Pin(),
				"psu2.p_out_raw":    ps2.PoutRaw(),
				"psu2.p_in_raw":     ps2.PinRaw(),
				"psu2.p_mode_raw":   ps2.ModeRaw(),
				"psu2.pmbus_rev":    ps2.PMBusRev(),
				"psu2.status_fans":  ps2.StatusFans(),
				"psu2.temperature1": ps2.Temp1(),
				"psu2.temperature2": ps2.Temp2(),
				"psu2.fan_speed":    ps2.FanSpeed(),
			},
		},
		&Info{
			name:     "temperature",
			prefixes: []string{"temperature."},
			attrs: machined.Attrs{
				"temperature.bmc_cpu":   cpu.ReadTemp(),
				"temperature.fan_front": hw.FrontTemp(),
				"temperature.fan_rear":  hw.RearTemp(),
				"temperature.pcb_board": 28.6,
			},
		},
	)

	tckr = time.NewTicker(time.Duration(redis_interval) * time.Second)
	go timerLoop()
	return nil
}

func timerLoop() {
	defer tckr.Stop()
	for {
		select {
		case <-tckr.C:
			timerIsr()
		}
	}
}

func updateUint16(v uint16, k string) {
	if v != RedisEnvShadow[k] {
		info.Publish(k, v)
		RedisEnvShadow[k] = v
	}
}

func updateFloat64(v float64, k string) {
	if v != RedisEnvShadow[k] {
		info.Publish(k, v)
		RedisEnvShadow[k] = v
	}
}

func updateString(v string, k string) {
	if v != RedisEnvShadow[k] {
		info.Publish(k, v)
		RedisEnvShadow[k] = v
	}
}

func force_sync(i uint16) {
	go timerIsr()
}

func change_interval(i uint16) {
	redis_interval = i
	tckr = time.NewTicker(time.Duration(redis_interval) * time.Second)
}

func clear_counters(s string) {
	t := time.Now()
	s = fmt.Sprintln(t.Format(time.RFC1123))
	s = s[:len(s)-1]
	redis_clear = s
}

func init_sync() uint16 {
	redis_sync = 0
	return redis_sync
}

func init_interval() uint16 {
	redis_interval = redis_default_interval
	return redis_interval
}

func init_counters() string {
	t := time.Now()
	s := fmt.Sprintln(t.Format(time.RFC1123))
	redis_clear = s[:len(s)-1]
	return redis_clear
}

func timerIsr() {
	timerMutex.Lock()
	defer timerMutex.Unlock()

	if stageFlagString == 1 {
		if _, ok := regWriteString[stageKeyString]; ok {
			regWriteString[stageKeyString](string(stageString))
		}
		stageFlagString = 0
	}
	if stageFlagInt == 1 {
		if _, ok := regWriteInt[stageKeyInt]; ok {
			regWriteInt[stageKeyInt](int(stageInt))
		}
		stageFlagInt = 0
	}
	if stageFlagUint16 == 1 {
		if _, ok := regWriteUint16[stageKeyUint16]; ok {
			regWriteUint16[stageKeyUint16](uint16(stageUint16))
		}
		stageFlagUint16 = 0
	}
	if stageFlagFloat64 == 1 {
		if _, ok := regWriteFloat64[stageKeyFloat64]; ok {
			regWriteFloat64[stageKeyFloat64](float64(stageFloat64))
		}
		stageFlagFloat64 = 0
	}

	updateUint16(redis_interval, "update.interval")
	updateString(redis_clear, "update.counters")

	updateFloat64(pm.Vout(1), "vmon.5v.sb")
	updateFloat64(pm.Vout(2), "vmon.3v8.bmc")
	updateFloat64(pm.Vout(3), "vmon.3v3.sys")
	updateFloat64(pm.Vout(4), "vmon.3v3.bmc")
	updateFloat64(pm.Vout(5), "vmon.3v3.sb")
	updateFloat64(pm.Vout(6), "vmon.1v0.thc")
	updateFloat64(pm.Vout(7), "vmon.1v8.sys")
	updateFloat64(pm.Vout(8), "vmon.1v25.sys")
	updateFloat64(pm.Vout(9), "vmon.1v2.ethx")
	updateFloat64(pm.Vout(10), "vmon.1v0.tha")

	updateUint16(hw.FanCount(1), "fan_tray.1.1.rpm")
	updateUint16(hw.FanCount(2), "fan_tray.1.2.rpm")
	updateUint16(hw.FanCount(3), "fan_tray.2.1.rpm")
	updateUint16(hw.FanCount(4), "fan_tray.2.2.rpm")
	updateUint16(hw.FanCount(5), "fan_tray.3.1.rpm")
	updateUint16(hw.FanCount(6), "fan_tray.3.2.rpm")
	updateUint16(hw.FanCount(7), "fan_tray.4.1.rpm")
	updateUint16(hw.FanCount(8), "fan_tray.4.2.rpm")

	updateString(ps1.PsuStatus(), "psu1.status")
	updateString(ps1.GetAdminState(), "psu1.admin.state")
	updateString(ps2.PsuStatus(), "psu2.status")
	updateString(ps2.GetAdminState(), "psu2.admin.state")

	updateString(ps1.MfgIdent(), "psu1.mfg_id")
	updateString(ps1.MfgModel(), "psu1.mfg_model")
	updateString(ps2.MfgIdent(), "psu2.mfg_id")
	updateString(ps2.MfgModel(), "psu2.mfg_model")

	updateUint16(ps1.Page(), "psu1.page")
	updateUint16(ps1.StatusWord(), "psu1.status_word")
	updateUint16(ps1.StatusVout(), "psu1.status_vout")
	updateUint16(ps1.StatusIout(), "psu1.status_iout")
	updateUint16(ps1.StatusInput(), "psu1.status_input")
	updateUint16(ps1.Iin(), "psu1.i_in")
	updateUint16(ps1.Iout(), "psu1.i_out")
	updateUint16(ps1.StatusTemp(), "psu1.status_temp")
	updateUint16(ps1.PMBusRev(), "psu1.pmbus_rev")
	updateUint16(ps1.StatusFans(), "psu1.status_fans")
	updateUint16(ps1.Temp1(), "psu1.temperature1")
	updateUint16(ps1.Temp2(), "psu1.temperature2")
	updateUint16(ps1.FanSpeed(), "psu1.fan_speed")
	updateFloat64(ps1.Vin(), "psu1.v_in")
	updateFloat64(ps1.Vout(), "psu1.v_out")
	updateFloat64(ps1.Pout(), "psu1.p_out")
	updateFloat64(ps1.Pin(), "psu1.p_in")
	updateUint16(ps1.PoutRaw(), "psu1.p_out_raw")
	updateUint16(ps1.PinRaw(), "psu1.p_in_raw")

	updateUint16(ps2.Page(), "psu2.page")
	updateUint16(ps2.StatusWord(), "psu2.status_word")
	updateUint16(ps2.StatusVout(), "psu2.status_vout")
	updateUint16(ps2.StatusIout(), "psu2.status_iout")
	updateUint16(ps2.StatusInput(), "psu2.status_input")
	updateUint16(ps2.Iin(), "psu2.i_in")
	updateUint16(ps2.Iout(), "psu2.i_out")
	updateUint16(ps2.StatusTemp(), "psu2.status_temp")
	updateUint16(ps2.PMBusRev(), "psu2.pmbus_rev")
	updateUint16(ps2.StatusFans(), "psu2.status_fans")
	updateUint16(ps2.Temp1(), "psu2.temperature1")
	updateUint16(ps2.Temp2(), "psu2.temperature2")
	updateUint16(ps2.FanSpeed(), "psu2.fan_speed")
	updateFloat64(ps2.Vin(), "psu2.v_in")
	updateFloat64(ps2.Vout(), "psu2.v_out")
	updateFloat64(ps2.Pout(), "psu2.p_out")
	updateFloat64(ps2.Pin(), "psu2.p_in")
	updateUint16(ps2.PoutRaw(), "psu2.p_out_raw")
	updateUint16(ps2.PinRaw(), "psu2.p_in_raw")

	updateFloat64(cpu.ReadTemp(), "temperature.bmc_cpu")
	updateFloat64(hw.FrontTemp(), "temperature.fan_front")
	updateFloat64(hw.RearTemp(), "temperature.fan_rear")

	updateString(fanTray.FanTrayStatus(1), "fan_tray.1.status")
	updateString(fanTray.FanTrayStatus(2), "fan_tray.2.status")
	updateString(fanTray.FanTrayStatus(3), "fan_tray.3.status")
	updateString(fanTray.FanTrayStatus(4), "fan_tray.4.status")
	updateString(hw.GetFanSpeed(), "fan_tray.speed")

	//updates front panel led state to reflect system state
	ledfp.LedStatus()
}

func (p *Info) Main(...string) error {
	for _, entry := range []struct{ name, unit string }{
		{"fan", "% max speed"},
		{"vmon", "volts"},
		{"temperature", "°C"},
	} {
		info.Publish("unit."+entry.name, entry.unit)
	}
	for k, a := range p.attrs {
		info.Publish(k, a)
		RedisEnvShadow[k] = a
	}
	return nil
}

func (*Info) Close() error {
	return nil
}

func (p *Info) Del(key string) error {
	if _, found := p.attrs[key]; !found {
		return info.CantDel(key)
	}
	delete(p.attrs, key)
	info.Publish("delete", key)
	return nil
}

func (p *Info) Prefixes(prefixes ...string) []string {
	if len(prefixes) > 0 {
		p.prefixes = prefixes
	}
	return p.prefixes
}

// this function is called by redis, do not do lengthy i2c calls here
func (p *Info) Set(key, value string) error {
	a, found := p.attrs[key]
	if !found {
		return info.CantSet(key)
	}
	switch t := a.(type) {
	case string:
		p.attrs[key] = value
		stageString = value
		stageKeyString = key
		stageFlagString = 1
		RedisEnvShadow[key] = value
	case int:
		i, err := strconv.ParseInt(value, 0, 0)
		if err != nil {
			return err
		}
		p.attrs[key] = i
		stageInt = int(i)
		stageKeyInt = key
		stageFlagInt = 1
		RedisEnvShadow[key] = int(i)
	case uint16:
		u, err := strconv.ParseUint(value, 0, 16)
		if err != nil {
			return err
		}
		p.attrs[key] = uint16(u)
		stageUint16 = uint16(u)
		stageKeyUint16 = key
		stageFlagUint16 = 1
		RedisEnvShadow[key] = uint16(u)
	case float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		p.attrs[key] = f
		stageFloat64 = float64(f)
		stageKeyFloat64 = key
		stageFlagFloat64 = 1
		RedisEnvShadow[key] = float64(f)
	default:
		if method, found := t.(parser); found {
			if err := method.Parse(value); err != nil {
				return err
			}
		} else {
			return info.CantSet(key)
		}
	}
	info.Publish(key, fmt.Sprint(p.attrs[key]))
	return nil
}

func (p *Info) String() string { return p.name }
