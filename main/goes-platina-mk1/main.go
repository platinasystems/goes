// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Platina's Mk1 TOR
package main

import (
	"fmt"
	"os"

	"github.com/platinasystems/fe1"
	platinasystemsGo "github.com/platinasystems/go"
	ipLinkAddType "github.com/platinasystems/go/goes/cmd/ip/link/add/type"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/basic"
	"github.com/platinasystems/go/internal/machine"
)

func main() {
	var ecode int
	machine.Name = "platina-mk1"
	platinasystemsGo.Packages = func() []map[string]string {
		return fe1.Packages
	}
	ipLinkAddType.Goes.ByName[machine.Name] = basic.Command("platina-mk1")
	if err := Goes.Main(os.Args...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		ecode = 1
	}
	os.Exit(ecode)
}
