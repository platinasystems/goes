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
	"github.com/platinasystems/go/internal/machine"
)

const name = "platina-mk1"

func main() {
	var ecode int
	machine.Name = name
	platinasystemsGo.Packages = func() []map[string]string {
		return fe1.Packages
	}
	if err := Goes.Main(os.Args...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		ecode = 1
	}
	os.Exit(ecode)
}
