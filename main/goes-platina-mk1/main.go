// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Platina's Mk1 TOR
package main

import (
	"fmt"
	"os"

	platinasystemsGo "github.com/platinasystems/go"
	"github.com/platinasystems/go/platform/mk1"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/plugins/fe1"
)

const name = "platina-mk1"

func main() {
	platinasystemsGo.Packages = func() []map[string]string {
		return fe1.Packages()
	}
	if err := mk1.Start(name); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
