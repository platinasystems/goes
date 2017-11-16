// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
)

// Goes is set by -test.goes
var Goes bool

func init() {
	flag.BoolVar(&Goes, "test.goes", false,
		"run goes command instead of test(s)")
}

// Exec will execuute given Goes().main with os.Args stripped of the
// goes-MACHINE.test program name and any leading -test/* arguments.
// This exits 0 if main returns nil; otherwise, it prints the error
// and exits 1.
//
// Usage:
//
//	func Test(t *testing.T) {
//		if Goes {
//			Exec(main.Goes().Main)
//		}
//		t.Run("Test1", func(t *testing.T) {
//			...
//		})
//		...
//	}
func Exec(main func(...string) error) {
	args := os.Args[1:]
	n := 0
	for ; n < len(args) && strings.HasPrefix(args[n], "-test."); n++ {
	}
	if n > 0 {
		copy(args[:len(args)-n], args[n:])
		args = args[:len(args)-n]
	}
	ecode := 0
	if err := main(args...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		ecode = 1
	}
	syscall.Exit(ecode)
}
