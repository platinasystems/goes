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

	"sort"
)

type MaskedStringer interface {
	MaskedString(mask MaskedStringer) string
}

type masked_string_pair struct{ v, m MaskedStringer }
type MaskedStrings struct {
	m map[string]masked_string_pair
}

func (x *MaskedStrings) Add(key string, v, m MaskedStringer) {
	if x.m == nil {
		x.m = make(map[string]masked_string_pair)
	}
	x.m[key] = masked_string_pair{v: v, m: m}
}

func (x *MaskedStrings) String() (s string) {
	type t struct{ k, v string }
	var ts []t
	for k, v := range x.m {
		ts = append(ts, t{k: k, v: v.v.MaskedString(v.m)})
	}
	sort.Slice(ts, func(i, j int) bool { return ts[i].k < ts[j].k })
	for i := range ts {
		t := &ts[i]
		if s != "" {
			s += ", "
		}
		s += t.k + ": " + t.v
	}
	return
}

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
