// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build amd64 arm

// This is an example goes machine.
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/platinasystems/go/builtinutils"
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/coreutils"
	"github.com/platinasystems/go/diagutils/dlv"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/machined"
	"github.com/platinasystems/go/netutils"
	"github.com/platinasystems/go/netutils/telnetd"
)

type parser interface {
	Parse(string) error
}

type ExampleInfo struct {
	name     string
	prefixes []string
	attrs    machined.Attrs
}

func main() {
	command.Plot(builtinutils.New()...)
	command.Plot(coreutils.New()...)
	command.Plot(dlv.New()...)
	command.Plot(netutils.New())
	command.Plot(telnetd.New())
	machined.Hook = func() {
		machined.NetLink.Prefixes("lo.", "eth0.")
		machined.InfoProviders = append(machined.InfoProviders,
			&ExampleInfo{
				name:     "current",
				prefixes: []string{"current."},
				attrs: machined.Attrs{
					"current.somewhere": 3.33,
				},
			},
			&ExampleInfo{
				name:     "fan",
				prefixes: []string{"fan."},
				attrs: machined.Attrs{
					"fan.front": 100,
					"fan.back":  100,
				},
			},
			&ExampleInfo{
				name:     "psu",
				prefixes: []string{"psu."},
				attrs: machined.Attrs{
					"psu.0": 5.01,
					"psu.1": 4.98,
				},
			},
			&ExampleInfo{
				name:     "potential",
				prefixes: []string{"potential."},
				attrs: machined.Attrs{
					"potential.1.8": 1.82,
					"potential.2.5": 2.53,
					"potential.5":   5.05,
					"potential.12":  11.98,
				},
			},
			&ExampleInfo{
				name:     "chassis",
				prefixes: []string{"slot."},
				attrs: machined.Attrs{
					"slot.0": "empty",
					"slot.1": "empty",
					"slot.2": "empty",
					"slot.3": "empty",
				},
			},
			&ExampleInfo{
				name:     "temperature",
				prefixes: []string{"temperature."},
				attrs: machined.Attrs{
					"temperature.cpu": 28.6,
				},
			})
	}
	os.Setenv("REDISD_DEVS", "lo eth0")
	err := goes.Main(os.Args...)
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "%s: %v\n", command.Prog, err)
		os.Exit(1)
	}
}

func (p *ExampleInfo) Main(...string) error {
	machined.Publish("machine", "example")
	for _, entry := range []struct{ name, unit string }{
		{"current", "milliamperes"},
		{"fan", "% max speed"},
		{"potential", "volts"},
		{"temperature", "°C"},
	} {
		machined.Publish("unit."+entry.name, entry.unit)
	}
	for k, a := range p.attrs {
		machined.Publish(k, a)
	}
	return nil
}

func (*ExampleInfo) Close() error {
	return nil
}

func (p *ExampleInfo) Del(key string) error {
	if _, found := p.attrs[key]; !found {
		return machined.CantDel(key)
	}
	delete(p.attrs, key)
	machined.Publish("delete", key)
	return nil
}

func (p *ExampleInfo) Prefixes(prefixes ...string) []string {
	if len(prefixes) > 0 {
		p.prefixes = prefixes
	}
	return p.prefixes
}

func (p *ExampleInfo) Set(key, value string) error {
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

func (p *ExampleInfo) String() string {
	return p.name
}
