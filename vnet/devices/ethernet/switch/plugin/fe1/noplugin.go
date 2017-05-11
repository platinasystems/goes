// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build noplugin

package fe1

import (
	"github.com/platinasystems/fe1"
	firmware "github.com/platinasystems/firmware-fe1a"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
)

func Packages() []map[string]string {
	return []map[string]string{fe1.Package, firmware.Package}
}

func AddPlatform(v *vnet.Vnet, ver int, nmacs uint32, basea ethernet.Address,
	init func(), leden func() error) {
	fe1.AddPlatform(v, ver, nmacs, basea, init, leden)
}

func Init(v *vnet.Vnet) { fe1.Init(v) }
