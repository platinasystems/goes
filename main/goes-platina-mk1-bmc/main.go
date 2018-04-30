// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This is the Baseboard Management Controller of Platina's Mk1 TOR.
package main

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/internal/machine"
)

const name = "platina-mk1-bmc"

func main() {
	var ecode int
	machine.Name = name
	if err := Goes.Main(os.Args...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		ecode = 1
	}
	os.Exit(ecode)
}
