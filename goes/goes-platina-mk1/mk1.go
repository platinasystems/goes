// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/builtinutils"
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/coreutils"
	"github.com/platinasystems/go/diagutils"
	"github.com/platinasystems/go/diagutils/dlv"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/fsutils"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/initutils"
	"github.com/platinasystems/go/initutils/start"
	"github.com/platinasystems/go/kutils"
	"github.com/platinasystems/go/machined"
	"github.com/platinasystems/go/machined/info"
	"github.com/platinasystems/go/machined/info/cmdline"
	"github.com/platinasystems/go/machined/info/hostname"
	"github.com/platinasystems/go/machined/info/netlink"
	"github.com/platinasystems/go/machined/info/uptime"
	"github.com/platinasystems/go/machined/info/version"
	"github.com/platinasystems/go/netutils"
	"github.com/platinasystems/go/netutils/npu"
	"github.com/platinasystems/go/redisutils"
	"github.com/platinasystems/go/sockfile"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/bus/pci"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"

	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const Machine = "platina-mk1"

type platform struct {
	vnet.Package
	*bcm.Platform
	p *Info
}

type parser interface {
	Parse(string) error
}

type Info struct {
	mutex    sync.Mutex
	name     string
	prefixes []string
	attrs    machined.Attrs
	v        *vnet.Vnet
	statsPoller
}

type AttrInfo struct {
	attr_name string
	attr      interface{}
}

var speedMap = map[string]float64{
	"100g": 100e9,
	"40g":  40e9,
	"10g":  10e9,
	"1g":   1e9,
}

// Would like to do "eth-0-0.speedsetting = 100e9,4" where 4 is number of lanes based off subport 0 here.
var portConfigs = []AttrInfo{
	{"speed", "100g"},
	{"autoneg", "false"},
	{"loopback", "false"},
}

func main() {
	command.Plot(builtinutils.New()...)
	command.Plot(coreutils.New()...)
	command.Plot(dlv.New()...)
	command.Plot(diagutils.New()...)
	command.Plot(fsutils.New()...)
	command.Plot(initutils.New()...)
	command.Plot(kutils.New()...)
	command.Plot(machined.New(), npu.New())
	command.Plot(netutils.New()...)
	command.Plot(redisutils.New()...)
	command.Sort()
	start.Hook = func() error {
		os.Setenv("REDISD", "lo eth0")
		return nil
	}
	machined.Hook = hook
	goes.Main()
}

func hook() error {
	machined.Plot(
		cmdline.New(),
		hostname.New(),
		netlink.New(),
		uptime.New(),
		version.New(),
		&Info{
			name:     "vnet",
			prefixes: []string{"eth-", "dp-"},
			attrs:    make(machined.Attrs),
		},
	)
	machined.Info["netlink"].Prefixes("lo.", "eth0.")
	return nil
}

func (p *Info) String() string { return p.name }

func (p *Info) Main(...string) error {
	info.Publish("machine", "platina-mk1")
	for _, entry := range []struct{ name, unit string }{
		{"current", "milliamperes"},
		{"fan", "% max speed"},
		{"potential", "volts"},
		{"temperature", "°C"},
	} {
		info.Publish("unit."+entry.name, entry.unit)
	}

	for port := 0; port < 32; port++ {
		for subport := 0; subport < 4; subport++ {
			// Initially only config subport 0 to match default
			if subport == 0 {
				for i := range portConfigs {
					k := fmt.Sprintf("eth-%02d-%d.%s", port, subport,
						portConfigs[i].attr_name)
					p.attrs[k] = portConfigs[i].attr
					// Publish configuration redis nodes
					info.Publish(k, fmt.Sprint(p.attrs[k]))
				}
			}
		}
	}

	gdbWait()

	return p.startVnet()
}

var oingoes int

func gdbWait() {
	// Change false to true to enable.
	// In gdb say "p 'main.oingoes'=1" to break out of loop.
	for false && oingoes == 0 {
		time.Sleep(1 * time.Second)
	}
}

func (p *Info) Close() error {
	// Stop vnet.
	return nil
}

func (p *Info) Del(key string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

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

// Send message to hw channel
func (p *Info) setHw(key, value string) (err error) {
	// If previous configuration existed on this
	// port, delete and start again.
	// Give new setting to vnet
	keyStr := strings.SplitN(key, ".", 2)
	switch keyStr[1] {
	case "speed":
		sp, ok := speedMap[value]
		if ok {
			err = p.SetSpeedHwIf(keyStr[0], sp)
		} else {
			fmt.Printf("Invalid speed setting for %s: %s\n", key, value)
		}
	case "autoneg":
		if value == "false" {
			sp, _ := speedMap[p.attrs[keyStr[0]+".speed"].(string)]
			err = p.SetSpeedHwIf(keyStr[0], sp)
		} else { // "true"
			err = p.SetSpeedHwIf(keyStr[0], 0)
		}
	}
	return
}

func (p *Info) settableKey(key string) error {
	var (
		found bool
	)
	keyStr := strings.SplitN(key, ".", 2)
	for i := range portConfigs {
		if portConfigs[i].attr_name == keyStr[1] {
			found = true
			break
		}
	}
	if !found {
		return info.CantSet(key)
	}
	return nil
}

func (p *Info) SetSpeedHwIf(hwif_name string, bandwidth float64) (err error) {
	hi, ok := p.v.HwIfByName(hwif_name)
	if !ok {
		err = fmt.Errorf("%s: hwif not found", hwif_name)
		return
	}
	err = hi.SetSpeed(p.v, vnet.Bandwidth(bandwidth))
	return
}

func (p *Info) Set(key, value string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	a, found := p.attrs[key]
	if !found {
		return info.CantSet(key)
	}

	// Test if this attribute is settable.
	errPerm := p.settableKey(key)
	if errPerm != nil {
		return errPerm
	}

	// Parse key to find port/subport and attribute
	// and send value down to the driver for validation.
	// If all good i.e. hw has set the value, publish it
	errHw := p.setHw(key, value)
	if errHw != nil {
		return errHw
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
	case uint64:
		u64, err := strconv.ParseUint(value, 0, 64)
		if err != nil {
			return err
		}
		p.attrs[key] = u64
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
			return info.CantSet(key)
		}
	}
	info.Publish(key, fmt.Sprint(p.attrs[key]))
	return nil
}

func (p *Info) startVnet() error {
	var in parse.Input
	in.Add("cli { listen { no-prompt socket " + sockfile.Path("npu") + "} }")
	v := &vnet.Vnet{}
	p.v = v

	bcm.Init(v)
	ethernet.Init(v)
	ip4.Init(v)
	ip6.Init(v)
	// Temporarily remove to get goesdeb running vnet
	//ixge.Init(v)
	pci.Init(v)
	pg.Init(v)
	unix.Init(v)

	plat := &platform{p: p}
	v.AddPackage("platform", plat)
	plat.DependsOn("pci-discovery") // after pci discovery

	p.statsPoller.p = p // initialize back pointer

	return v.Run(&in)
}

const statsPollerInterval = 5

type statsPoller struct {
	vnet.Event
	p *Info
}

func (p *statsPoller) startStatsPoller() { p.p.v.AddTimedEvent(p, statsPollerInterval) }
func (p *statsPoller) String() string    { return "redis stats poller" }
func (p *statsPoller) EventAction() {
	const (
		includeZeroCounters = false
		unixInterfacesOnly  = true // only front panel ports (e.g. no bcm-cpu or loopback ports)
	)
	p.p.v.ForeachHwIfCounter(includeZeroCounters, unixInterfacesOnly,
		func(hi vnet.Hi, counter string, count uint64) {
			info.Publish(hi.Name(p.p.v)+"."+strings.Replace(counter, " ", "_", -1), count)
		})
	p.startStatsPoller() // schedule next event
}
