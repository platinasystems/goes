// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import "github.com/platinasystems/go/goes/cmd/daemons"

func init() {
	daemons.Init = [][]string{
		[]string{"redisd"},
		[]string{"fantrayd"},
		[]string{"fspd"},
		[]string{"i2cd"},
		[]string{"imx6d"},
		[]string{"ledgpiod"},
		[]string{"mmclogd"},
		[]string{"upgraded"},
		[]string{"uptimed"},
		[]string{"telnetd"},
		[]string{"ucd9090d"},
		[]string{"w83795d"},
	}
}
