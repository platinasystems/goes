// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Platina's Mk1 TOR
package main

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/platform/mk1"
)

const name = "platina-mk1"

func main() {
	if err := mk1.Start(name); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
