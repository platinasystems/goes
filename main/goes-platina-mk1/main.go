// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Platina's Mk1 TOR
package main

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/internal/machine"
)

func main() {
	var ecode int
	machine.Name = "platina-mk1"
	if err := Goes.Main(os.Args...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		ecode = 1
	}
	os.Exit(ecode)
}
