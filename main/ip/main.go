// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build amd64 arm

// This is a reimplementation of iproute2.
package main

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/goes/cmd/ip"
)

const name = "ip"

func main() {
	var ecode int
	if err := ip.Goes.Main(os.Args...); err != nil {
		fmt.Fprintln(os.Stderr, "ip:", err)
		ecode = 1
	}
	os.Exit(ecode)
}
