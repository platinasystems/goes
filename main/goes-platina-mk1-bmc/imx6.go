// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import "github.com/platinasystems/go/internal/goes/cmd/imx6"

func init() { imx6.Init = imx6Init }

func imx6Init() {
	imx6.VpageByKey = map[string]uint8{
		"bmc.temperature.units.C": 1,
	}
}
