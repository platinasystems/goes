// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build amd64 arm

// This is the goes installer that extracts a zipped machine.
package main

import (
	"fmt"
	"io"
	"os"
)

var Exit = os.Exit
var Stderr io.Writer = os.Stderr

func main() {
	err := Goes().Main()
	if err != nil {
		fmt.Fprintln(Stderr, err)
		Exit(1)
	}
}
