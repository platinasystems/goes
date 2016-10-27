// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package example

import (
	"fmt"
	"os"
	"strconv"

	"github.com/platinasystems/go/machined"
	"github.com/platinasystems/go/machined/info"
	"github.com/platinasystems/go/machined/info/cmdline"
	"github.com/platinasystems/go/machined/info/hostname"
	"github.com/platinasystems/go/machined/info/netlink"
	"github.com/platinasystems/go/machined/info/uptime"
	"github.com/platinasystems/go/machined/info/version"
)

var GoesdHook = SbinInitHook

func SbinInitHook() error {
	os.Setenv("REDISD", "lo eth0")
	return nil
}

func MachineHook() error {
	machined.Plot(
		cmdline.New(),
		hostname.New(),
		netlink.New(),
		uptime.New(),
		version.New(),
		&Example{
			name:     "current",
			prefixes: []string{"current."},
			attrs: machined.Attrs{
				"current.somewhere": 3.33,
			},
		},
		&Example{
			name:     "fan",
			prefixes: []string{"fan."},
			attrs: machined.Attrs{
				"fan.front": 100,
				"fan.back":  100,
			},
		},
		&Example{
			name:     "psu",
			prefixes: []string{"psu."},
			attrs: machined.Attrs{
				"psu.0": 5.01,
				"psu.1": 4.98,
			},
		},
		&Example{
			name:     "potential",
			prefixes: []string{"potential."},
			attrs: machined.Attrs{
				"potential.1.8": 1.82,
				"potential.2.5": 2.53,
				"potential.5":   5.05,
				"potential.12":  11.98,
			},
		},
		&Example{
			name:     "chassis",
			prefixes: []string{"slot."},
			attrs: machined.Attrs{
				"slot.0": "empty",
				"slot.1": "empty",
				"slot.2": "empty",
				"slot.3": "empty",
			},
		},
		&Example{
			name:     "temperature",
			prefixes: []string{"temperature."},
			attrs: machined.Attrs{
				"temperature.cpu": 28.6,
			},
		},
	)
	machined.Info["netlink"].Prefixes("lo.", "eth0.")
	return nil
}

type parser interface {
	Parse(string) error
}

type Example struct {
	name     string
	prefixes []string
	attrs    machined.Attrs
}

func (p *Example) String() string { return p.name }

func (p *Example) Main(...string) error {
	info.Publish("machine", "example")
	for _, entry := range []struct{ name, unit string }{
		{"current", "milliamperes"},
		{"fan", "% max speed"},
		{"potential", "volts"},
		{"temperature", "°C"},
	} {
		info.Publish("unit."+entry.name, entry.unit)
	}
	for k, a := range p.attrs {
		info.Publish(k, a)
	}
	return nil
}

func (*Example) Close() error {
	return nil
}

func (p *Example) Del(key string) error {
	if _, found := p.attrs[key]; !found {
		return info.CantDel(key)
	}
	delete(p.attrs, key)
	info.Publish("delete", key)
	return nil
}

func (p *Example) Prefixes(prefixes ...string) []string {
	if len(prefixes) > 0 {
		p.prefixes = prefixes
	}
	return p.prefixes
}

func (p *Example) Set(key, value string) error {
	a, found := p.attrs[key]
	if !found {
		return info.CantSet(key)
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
			return info.CantSet(key)
		}
	}
	info.Publish(key, fmt.Sprint(p.attrs[key]))
	return nil
}
