// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/goes/cmd/tempd"
)

func init() {
	tempd.Init = func() {
		tempd.VpageByKey = map[string]uint8{
			"sys.cpu.coretemp.units.C": 0,
			"bmc.redis.status":         0,
		}
	}
}
