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
	"github.com/platinasystems/go/initutils/goesd"
	"github.com/platinasystems/go/kutils"
	"github.com/platinasystems/go/machined"
	"github.com/platinasystems/go/netutils"
	"github.com/platinasystems/go/redisutils"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/bus/pci"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm"

	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	Machine        = "platina-mk1"
	VnetCmdSock    = "/run/goes/socks/npu"
	statsTimerTick = 5
)

type platform struct {
	vnet.Package
	*bcm.Platform
}

type parser interface {
	Parse(string) error
}

type Info struct {
	mutex    sync.Mutex
	name     string
	prefixes []string
	attrs    machined.Attrs
}

type AttrInfo struct {
	attr_name string
	attr      interface{}
}

// Would like to do "eth-0-0.speedsetting = 100e9,4" where 4 is number of lanes based off subport 0 here.
var portConfigs = []AttrInfo{
	{"speedsetting", "100e9,4"},
	{"autoneg", "false"},
	{"loopback", "false"},
}

func main() {
	command.Plot(builtinutils.New()...)
	command.Plot(coreutils.New()...)
	command.Plot(dlv.New()...)
	command.Plot(diagutils.New()...)
	command.Plot(fsutils.New()...)
	command.Plot(goesd.New(), machined.New())
	command.Plot(kutils.New()...)
	command.Plot(netutils.New()...)
	command.Plot(redisutils.New()...)
	command.Sort()
	goesd.Hook = func() error {
		os.Setenv("REDISD", "lo eth0")
		return nil
	}
	machined.Hook = func() {
		machined.NetLink.Prefixes("lo.", "eth0.")
		machined.InfoProviders = append(machined.InfoProviders, &Info{
			name:     "mk1",
			prefixes: []string{"eth-", "dp-"},
			attrs:    make(machined.Attrs),
		})
	}
	goes.Main()
}

func (p *Info) Main(...string) error {
	machined.Publish("machine", "platina-mk1")
	for _, entry := range []struct{ name, unit string }{
		{"current", "milliamperes"},
		{"fan", "% max speed"},
		{"potential", "volts"},
		{"temperature", "°C"},
	} {
		machined.Publish("unit."+entry.name, entry.unit)
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
					machined.Publish(k, fmt.Sprint(p.attrs[k]))
				}
			}
		}
	}

	var in parse.Input
	vnetArgsLine := fmt.Sprint("cli { listen { socket ", VnetCmdSock, " no-prompt } }")
	vnetArgs := strings.Split(vnetArgsLine, " ")
	in.Add(vnetArgs[0:]...)
	v := &vnet.Vnet{}

	bcm.Init(v)
	ethernet.Init(v)
	ip4.Init(v)
	ip6.Init(v)
	// Temporarily remove to get goesdeb running vnet
	//ixge.Init(v)
	pci.Init(v)
	pg.Init(v)
	unix.Init(v)

	plat := &platform{}
	v.AddPackage("platform", plat)
	plat.DependsOn("pci-discovery") // after pci discovery

	// Set redis publish pointer so vnet can push updates
	//vnet.Publish = machined.Publish

	go func() {
		err := v.Run(&in)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
	}()

	initStatsTimer(v)
	return nil
}

func (p *Info) Close() error {
	return nil
}

func (p *Info) Del(key string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

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

// Send message to hw channel
func (p *Info) setHw(key, value string) error {
	// If previous configuration existed on this
	// port, delete and start again.

	// Send new setting to vnetdevices layer via channel

	return nil
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
		return machined.CantSet(key)
	}
	return nil
}

func (p *Info) Set(key, value string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	a, found := p.attrs[key]
	if !found {
		return machined.CantSet(key)
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
			return machined.CantSet(key)
		}
	}
	machined.Publish(key, fmt.Sprint(p.attrs[key]))
	return nil
}

func initStatsTimer(v *vnet.Vnet) {
	// Wait for vnet to setup server
	time.Sleep(2 * time.Second)

	go func() {
		// Start timer to poll npu counters (every 5 seconds)
		ticker := time.NewTicker(time.Second * statsTimerTick)

		for _ = range ticker.C {
			v.ForeachHwIfCounter(false, func(hi vnet.Hi, counter string, count uint64) {
				hiName := hi.Name(v)
				// Limit display to front-panel ports i.e.  "eth-*" ?
				countVal := fmt.Sprintf("%d", count)
				machined.Publish(strings.Replace(hiName+"."+counter, " ", "_", -1), countVal)
			})
		}
	}()
	return
}

func (p *Info) String() string { return p.name }
