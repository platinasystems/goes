// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build arm

// This is the Baseboard Management Controller of Platina's Mk2 Management
// Card.
package main

import (
	"fmt"
	"os"
)

func main() {
	var ecode int
	args := os.Args
	if os.Args[0] == "/init" {
		args = os.Args[1:]
	}
	if err := Goes.Main(args...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		ecode = 1
	}
	os.Exit(ecode)
}
