// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package gpio provides cli command to query/configure GPIO pins.
package gpio

import (
	"github.com/platinasystems/fdt"
	. "github.com/platinasystems/gpio"
	"github.com/platinasystems/oops"

	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type gpio struct{ oops.Id }

var Gpio = &gpio{"gpio"}

var File = "/boot/linux.dtb"
var GpioAlias GpioAliasMap
var Gpios PinMap

func (*gpio) Usage() string {
	return "gpio PIN_NAME [VALUE]"
}

// Build map of gpio pins for this gpio controller
func GatherGpioAliases(n *fdt.Node) {
	for p, pn := range n.Properties {
		if strings.Contains(p, "gpio") {
			val := strings.Split(string(pn), "\x00")
			v := strings.Split(val[0], "/")
			GpioAlias[p] = v[len(v)-1]
		}
	}
}

func buildPinMap(name string, mode string, bank string, index string) {
	i, _ := strconv.Atoi(index)
	Gpios[name] = GpioPinMode[mode] | GpioBankToBase[bank] |
		Pin(i)
}

// Build map of gpio pins for this gpio controller
func GatherGpioPins(n *fdt.Node, name string, value string) {
	var pn []string
	var mode string

	for na, al := range GpioAlias {
		if al == n.Name {
			for _, c := range n.Children {
				for p, _ := range c.Properties {
					switch p {
					case "gpio-pin-desc":
						pn = strings.Split(c.Name, "@")
					case "output-high", "output-low", "input":
						mode = p
					}
				}
				if mode != "" {
					buildPinMap(pn[0], mode, na, pn[1])
				}
				mode = ""
			}
		}
	}
}

var once sync.Once

type PinValue struct {
	Name  string
	PinNum  Pin
	Value bool
}
type pinValues []PinValue

func (p *PinValue) String() string {
	kind := "IN"
	if p.PinNum&IsOutputHi != 0 {
		kind = "OUT HI"
	}
	if p.PinNum&IsOutputLo != 0 {
		kind = "OUT LO"
	}
	return fmt.Sprintf("%8s%32s: %v", kind, p.Name, p.Value)
}

func GpioInit () {

	GpioAlias = make(GpioAliasMap)
        Gpios = make(PinMap)

        // Parse linux.dtb to generate gpio map for this machine
        if b, err := ioutil.ReadFile(File); err == nil {
                t := &fdt.Tree{Debug: false, IsLittleEndian: false}
                t.Parse(b)

                t.MatchNode("aliases", GatherGpioAliases)
                t.EachProperty("gpio-controller", "", GatherGpioPins)
        } else {
                fmt.Println(err)
        }

	// Set pin directions.
	for _, p := range Gpios {
		err := p.SetDirection()
		if err != nil {
		// Don't panic just report error and continue.
			fmt.Println(err)
		}
	}
	return
}


// Implement sort.Interface
func (p pinValues) Len() int           { return len(p) }
func (p pinValues) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p pinValues) Less(i, j int) bool { return p[i].Name < p[j].Name }

func (p *gpio) Main(args ...string) {
	if len(args) > 2 {
		p.Panic(args[2:], ": unexpected")
	}

	GpioAlias = make(GpioAliasMap)
	Gpios = make(PinMap)

	// Parse linux.dtb to generate gpio map for this machine
	if b, err := ioutil.ReadFile(File); err == nil {
		t := &fdt.Tree{Debug: false, IsLittleEndian: false}
		t.Parse(b)

		t.MatchNode("aliases", GatherGpioAliases)
		t.EachProperty("gpio-controller", "", GatherGpioPins)
	} else {
		p.Panic(File, ": ", err)
	}


	if len(args) == 1 && args[0] == "default" {
		// Set pin directions.
		for _, p := range Gpios {
			err := p.SetDirection()
			if err != nil {
				// Don't panic just report error and continue.
				fmt.Println(err)
			}
		}
		return
	}

	// No args?  Report all pin values.
	if len(args) == 0 {

		pvs := pinValues{}
		for n, p := range Gpios {
			v, err := p.Value()
			if err != nil {
				// Don't panic just report error and continue.
				fmt.Println(err)
			}
			pvs = append(pvs, PinValue{Name: n, Value: v, PinNum: p})
		}
		sort.Sort(pvs)
		for i := range pvs {
			fmt.Println(&pvs[i])
		}
		return
	}

	if len(args) > 0 {
		pv := PinValue{Name: args[0]}
		ok := false
		if pv.PinNum, ok = Gpios[pv.Name]; !ok {
			p.Panic("no such pin: " + pv.Name)
		}

		if len(args) == 2 {
			switch args[1] {
			case "true", "1":
				ok = true
			case "false", "0":
				ok = false
			default:
				p.Panic("expected true|false or 0|1, got ",
					args[1])
			}
			pv.PinNum.SetValue(ok)
		}

		// Single arg? Read specified pin value.
		if v, err := pv.PinNum.Value(); err != nil {
			fmt.Println(err)
		} else {
			pv.Value = v
		}
		fmt.Println(&pv)
	}
}
