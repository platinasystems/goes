// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"os"
	"strings"
	"syscall"
)

// Main runs the main function if given the "-test.main" flag.  With said flag,
// this strip os.Args[0] and any leading -test.* options and os.Exit(0) if the
// main returns.
func Main(main func()) {
	if !*IsMain {
		return
	}
	os.Args = os.Args[1:]
	for len(os.Args) > 0 && strings.HasPrefix(os.Args[0], "-test.") {
		os.Args = os.Args[1:]
	}
	main()
	syscall.Exit(0)
}
