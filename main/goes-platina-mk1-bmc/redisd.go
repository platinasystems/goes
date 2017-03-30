// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/internal/goes/cmd/eeprom/platina_eeprom"
	"github.com/platinasystems/go/internal/goes/cmd/redisd"
)

func init() {
	redisd.Init = func() {
		redisd.Machine = "platina-mk1-bmc"
		redisd.Devs = []string{"lo", "eth0"}
		redisd.Hook = platina_eeprom.RedisdHook
	}
}
