// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build arm

// This is an example Baseboard Management Controller.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/platinasystems/go/builtinutils"
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/coreutils"
	"github.com/platinasystems/go/diagutils"
	"github.com/platinasystems/go/emptych"
	"github.com/platinasystems/go/environ/nuvoton"
	"github.com/platinasystems/go/environ/nxp"
	"github.com/platinasystems/go/environ/ti"
	"github.com/platinasystems/go/fdt"
	"github.com/platinasystems/go/fdtgpio"
	"github.com/platinasystems/go/fsutils"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/gpio"
	"github.com/platinasystems/go/initutils/sbininit"
	"github.com/platinasystems/go/initutils/slashinit"
	"github.com/platinasystems/go/kutils"
	"github.com/platinasystems/go/machined"
	"github.com/platinasystems/go/netutils"
	"github.com/platinasystems/go/redisutils"
)

type parser interface {
	Parse(string) error
}

type Info struct {
	emptych.In
	name     string
	prefixes []string
	attrs    machined.Attrs
}

var RedisEnvShadow = map[string]interface{}{}

// Machine specific
const (
	w83795Bus     = 0
	w83795Adr     = 0x2f
	w83795MuxAdr  = 0x76
	w83795MuxVal  = 0x80
	ucd9090Bus    = 0
	ucd9090Adr    = 0x7e
	ucd9090MuxAdr = 0x76
	ucd9090MuxVal = 0x01
)

var hw = w83795.HwMonitor{w83795Bus, w83795Adr, w83795MuxAdr, w83795MuxVal}
var pm = ucd9090.PowerMon{ucd9090Bus, ucd9090Adr, ucd9090MuxAdr, ucd9090MuxVal}
var cpu = imx6.Cpu{}

func main() {
	gpio.File = "/boot/platina-mk1-bmc.dtb"
	command.Plot(builtinutils.New()...)
	command.Plot(coreutils.New()...)
	command.Plot(diagutils.New()...)
	command.Plot(fsutils.New()...)
	command.Plot(sbininit.New(), slashinit.New())
	command.Plot(kutils.New()...)
	command.Plot(netutils.New()...)
	command.Plot(redisutils.New()...)
	command.Sort()
	sbininit.Hook = func() error {
		os.Setenv("REDISD", "lo eth0")
		return nil
	}
	machined.Hook = func() {
		gpioInit()
		machined.NetLink.Prefixes("lo.", "eth0.")
		machined.InfoProviders = append(machined.InfoProviders,
			&Info{name: "platina-mk1-bmc"},
			&Info{
				name:     "fan",
				prefixes: []string{"fan."},
				attrs: machined.Attrs{
					"fan.front": 100,
					"fan.rear":  100,
				},
			},
			&Info{
				name:     "psu",
				prefixes: []string{"psu."},
				attrs: machined.Attrs{
					"psu.0": 5.01,
					"psu.1": 4.98,
				},
			},
			&Info{
				name:     "potential",
				prefixes: []string{"potential."},
				attrs: machined.Attrs{
					"potential.5":         pm.Vout(1),
					"potential.3.8.bmc":   pm.Vout(2),
					"potential.3.3":       pm.Vout(3),
					"potential.3.3.bmc":   pm.Vout(4),
					"potential.3.3.sys":   pm.Vout(5),
					"potential.2.5.sys":   pm.Vout(6),
					"potential.1.8":       pm.Vout(7),
					"potential.1.25":      pm.Vout(8),
					"potential.1.2":       pm.Vout(9),
					"potential.1.0.tom.a": pm.Vout(10),
				},
			},
			&Info{
				name:     "chassis",
				prefixes: []string{"fan_speed."},
				attrs: machined.Attrs{
					"fan_speed.1": hw.FanCount(1),
					"fan_speed.2": hw.FanCount(2),
					"fan_speed.3": hw.FanCount(3),
					"fan_speed.4": hw.FanCount(4),
					"fan_speed.5": hw.FanCount(5),
					"fan_speed.6": hw.FanCount(6),
					"fan_speed.7": hw.FanCount(7),
					"fan_speed.8": hw.FanCount(8),
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
			})
	}
	goes.Main()
}

func updateUint16(v uint16, k string) {
	if v != RedisEnvShadow[k] {
		machined.Publish(k, v)
		RedisEnvShadow[k] = v
	}
}

func updateFloat64(v float64, k string) {
	if v != RedisEnvShadow[k] {
		machined.Publish(k, v)
		RedisEnvShadow[k] = v
	}
}

func (p *Info) update() {
	updateFloat64(pm.Vout(1), "potential.5")
	updateFloat64(pm.Vout(2), "potential.3.8.bmc")
	updateFloat64(pm.Vout(3), "potential.3.3")
	updateFloat64(pm.Vout(4), "potential.3.3.bmc")
	updateFloat64(pm.Vout(5), "potential.3.3.sys")
	updateFloat64(pm.Vout(6), "potential.2.5.sys")
	updateFloat64(pm.Vout(7), "potential.1.8")
	updateFloat64(pm.Vout(8), "potential.1.25")
	updateFloat64(pm.Vout(9), "potential.1.2")
	updateFloat64(pm.Vout(10), "potential.1.0.tom.a")

	updateUint16(hw.FanCount(1), "fan_speed.1")
	updateUint16(hw.FanCount(2), "fan_speed.2")
	updateUint16(hw.FanCount(3), "fan_speed.3")
	updateUint16(hw.FanCount(4), "fan_speed.4")
	updateUint16(hw.FanCount(5), "fan_speed.5")
	updateUint16(hw.FanCount(6), "fan_speed.6")
	updateUint16(hw.FanCount(7), "fan_speed.7")
	updateUint16(hw.FanCount(8), "fan_speed.8")

	updateFloat64(cpu.ReadTemp(), "temperature.bmc_cpu")
	updateFloat64(hw.FrontTemp(), "temperature.fan_front")
	updateFloat64(hw.RearTemp(), "temperature.fan_rear")
}

func (p *Info) Main(...string) error {
	machined.Publish("machine", "platina-mk1-bmc")
	for _, entry := range []struct{ name, unit string }{
		{"fan", "% max speed"},
		{"potential", "volts"},
		{"temperature", "°C"},
	} {
		machined.Publish("unit."+entry.name, entry.unit)
	}
	for k, a := range p.attrs {
		machined.Publish(k, a)
		RedisEnvShadow[k] = a
	}

	stop := emptych.Make()
	p.In = emptych.In(stop)
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return nil
		case <-t.C:
			p.update()
		}
	}
	return nil
}

func (*Info) Close() error {
	return nil
}

func (p *Info) Del(key string) error {
	if _, found := p.attrs[key]; !found {
		return machined.CantDel(key)
	}
	delete(p.attrs, key)
	machined.Publish("delete", key)
	return nil
}

func (p *Info) Prefixes(prefixes ...string) []string {
	if len(prefixes) > 0 {
		p.prefixes = prefixes
	}
	return p.prefixes
}

func (p *Info) Set(key, value string) error {
	a, found := p.attrs[key]
	if !found {
		return machined.CantSet(key)
	}
	switch t := a.(type) {
	case string:
		p.attrs[key] = value
	case int:
		i, err := strconv.ParseInt(value, 0, 0)
		if err != nil {
			return err
		}
		p.attrs[key] = i
	case float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		p.attrs[key] = f
	default:
		if method, found := t.(parser); found {
			if err := method.Parse(value); err != nil {
				return err
			}
		} else {
			return machined.CantSet(key)
		}
	}
	machined.Publish(key, fmt.Sprint(p.attrs[key]))
	return nil
}

func (p *Info) String() string { return p.name }

func gpioInit() {
	gpio.Aliases = make(gpio.GpioAliasMap)
	gpio.Pins = make(gpio.PinMap)

	// Parse linux.dtb to generate gpio map for this machine
	if b, err := ioutil.ReadFile(gpio.File); err == nil {
		t := &fdt.Tree{Debug: false, IsLittleEndian: false}
		t.Parse(b)

		t.MatchNode("aliases", fdtgpio.GatherAliases)
		t.EachProperty("gpio-controller", "", fdtgpio.GatherPins)
	} else {
		fmt.Println(err)
	}

	// Set pin directions.
	for _, p := range gpio.Pins {
		err := p.SetDirection()
		if err != nil {
			// Don't panic just report error and continue.
			fmt.Println(err)
		}
	}
	return
}
