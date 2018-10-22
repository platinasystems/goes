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

	"github.com/platinasystems/redis"
)

const name = "platina-mk2-mc1"

func main() {
	var ecode int
	redis.DefaultHash = name
	if err := Goes.Main(os.Args...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		ecode = 1
	}
	os.Exit(ecode)
}
