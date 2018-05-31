// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cpu"
	"github.com/platinasystems/go/elib/dep"
	"github.com/platinasystems/go/elib/loop"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/internal/xeth"
)

var Xeth *xeth.Xeth

type PortEntry struct {
	Ifindex uint64
	Net     uint64
	Flags   xeth.EthtoolFlagBits
	Speed   xeth.Mbps
}

var Ports map[string]*PortEntry

func SetPort(ifname string) *PortEntry {
	if Ports == nil {
		Ports = make(map[string]*PortEntry)
	}
	entry, found := Ports[ifname]
	if !found {
		entry = new(PortEntry)
		Ports[ifname] = entry
	}
	return entry
}

var (
	PortIsCopper = func(ifname string) bool { return false }
	PortIsFec74  = func(ifname string) bool { return false }
	PortIsFec91  = func(ifname string) bool { return false }
)

type RxTx int

const (
	Rx RxTx = iota
	Tx
	NRxTx
)

var rxTxStrings = [...]string{
	Rx: "rx",
	Tx: "tx",
}

func (x RxTx) String() (s string) {
	return elib.Stringer(rxTxStrings[:], int(x))
}

//go:generate gentemplate -id initHook -d Package=vnet -d DepsType=initHookVec -d Type=initHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl
type initHook func(v *Vnet)

var initHooks initHookVec

func AddInit(f initHook, deps ...*dep.Dep) { initHooks.Add(f, deps...) }

func (v *Vnet) configure(in *parse.Input) (err error) {
	if err = v.ConfigurePackages(in); err != nil {
		return
	}
	if err = v.InitPackages(); err != nil {
		return
	}
	return
}
func (v *Vnet) TimeDiff(t0, t1 cpu.Time) float64 { return v.loop.TimeDiff(t1, t0) }

func (v *Vnet) Run(in *parse.Input) (err error) {
	loop.AddInit(func(l *loop.Loop) {
		v.interfaceMain.init()
		v.CliInit()
		v.eventInit()
		for i := range initHooks.hooks {
			initHooks.Get(i)(v)
		}
		if err := v.configure(in); err != nil {
			panic(err)
		}
	})
	v.loop.Run()
	err = v.ExitPackages()
	return
}

func (v *Vnet) Quit() { v.loop.Quit() }
