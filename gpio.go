// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package gpio provides cli command to query/configure GPIO pins.
package gpio

import (
	"github.com/platinasystems/fdt"
	"github.com/platinasystems/goes"
	. "github.com/platinasystems/gpio"
	"github.com/platinasystems/oops"

	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	dtbFile = "/boot/linux.dtb"
)

type Gpio struct{ oops.Id }

var gpioAlias GpioAliasMap
var gpios PinMap

func init() {
	goes.Command.Map(&Gpio{oops.Id("gpio")})

	// Parse linux.dtb to generate gpio map for this machine
	b, err := ioutil.ReadFile(dtbFile)
	if err != nil {
		panic(err)
	}
	t := &fdt.Tree{Debug: false, IsLittleEndian: false}
	t.Parse(b)

	gpioAlias = make(GpioAliasMap)
	gpios = make(PinMap)

	t.MatchNode("aliases", gatherGpioAliases)
	t.EachProperty("gpio-controller", "", gatherGpioPins)
}

func (p *Gpio) Usage() string {
	return "gpio PIN_NAME [VALUE]"
}

// Build map of gpio pins for this gpio controller
func gatherGpioAliases(n *fdt.Node) {
	for p, pn := range n.Properties {
		if strings.Contains(p, "gpio") {
			val := strings.Split(string(pn), "\x00")
			v := strings.Split(val[0], "/")
			gpioAlias[p] = v[len(v)-1]
		}
	}
}

func buildPinMap(name string, mode string, bank string, index string) {
	i, _ := strconv.Atoi(index)
	gpios[name] = GpioPinMode[mode] | GpioBankToBase[bank] |
		Pin(i)
}

// Build map of gpio pins for this gpio controller
func gatherGpioPins(n *fdt.Node, name string, value string) {
	var pn []string
	var mode string

	for na, al := range gpioAlias {
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

type pinValue struct {
	name  string
	pin   Pin
	value bool
}
type pinValues []pinValue

func (p *pinValue) String() string {
	kind := "IN"
	if p.pin&IsOutputHi != 0 {
		kind = "OUT HI"
	}
	if p.pin&IsOutputLo != 0 {
		kind = "OUT LO"
	}
	return fmt.Sprintf("%8s%32s: %v", kind, p.name, p.value)
}

// Implement sort.Interface
func (p pinValues) Len() int           { return len(p) }
func (p pinValues) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p pinValues) Less(i, j int) bool { return p[i].name < p[j].name }

func (p *Gpio) Main(ctx *goes.Context, args ...string) {
	if len(args) > 2 {
		p.Panic(args[2:], ": unexpected")
	}

	if len(args) == 1 && args[0] == "default" {
		// Set pin directions.
		for _, p := range gpios {
			err := p.SetDirection()
			if err != nil {
				// Don't panic just report error and continue.
				ctx.Println(err)
			}
		}
		return
	}

	// No args?  Report all pin values.
	if len(args) == 0 {

		pvs := pinValues{}
		for n, p := range gpios {
			v, err := p.Value()
			if err != nil {
				// Don't panic just report error and continue.
				ctx.Println(err)
			}
			pvs = append(pvs, pinValue{name: n, value: v, pin: p})
		}
		sort.Sort(pvs)
		for i := range pvs {
			ctx.Println(&pvs[i])
		}
		return
	}

	if len(args) > 0 {
		pv := pinValue{name: args[0]}
		ok := false
		if pv.pin, ok = gpios[pv.name]; !ok {
			p.Panic("no such pin: " + pv.name)
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
			pv.pin.SetValue(ok)
		}

		// Single arg? Read specified pin value.
		if v, err := pv.pin.Value(); err != nil {
			ctx.Println(err)
		} else {
			pv.value = v
		}
		ctx.Println(&pv)
	}
}
