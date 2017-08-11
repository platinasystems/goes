// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build amd64 arm

// This is a reimplementation of iproute2.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/platinasystems/go/goes/cmd/ip"
)

var Args = os.Args
var Exit = os.Exit
var Stderr io.Writer = os.Stderr

func main() {
	if err := ip.New().Main(Args...); err != nil {
		fmt.Fprintln(Stderr, "ip:", err)
		Exit(1)
	}
}
