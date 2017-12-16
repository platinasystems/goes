// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import "github.com/platinasystems/go/goes/cmd/daemons"

func daemonsInit() {
	daemons.Init = [][]string{
		[]string{"redisd"},
		[]string{"tempd"},
		[]string{"uptimed"},
		[]string{"vnetd"},
	}
}
