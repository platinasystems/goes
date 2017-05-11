// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build !noplugin

package fe1

import (
	"plugin"

	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
)

const FileName = "/usr/lib/goes/fe1.so"

var lib *plugin.Plugin

func lookup(name string) plugin.Symbol {
	if lib == nil {
		var err error
		lib, err = plugin.Open(FileName)
		if err != nil {
			panic(err)
		}
	}
	sym, err := lib.Lookup(name)
	if err != nil {
		panic(err)
	}
	return sym
}

func Packages() []map[string]string {
	return lookup("Packages").(func() []map[string]string)()
}

func Init(v *vnet.Vnet) {
	lookup("Init").(func(*vnet.Vnet))(v)
}

func AddPlatform(v *vnet.Vnet, ver int, nmacs uint32, basea ethernet.Address,
	init func(), leden func() error) {
	lookup("AddPlatform").(func(*vnet.Vnet, int, uint32, ethernet.Address,
		func(), func() error))(v, ver, nmacs, basea, init, leden)
}
