// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/fsutils"
)

func main() {
	command.Plot(fsutils.New()...)
	command.Sort()
	if len(os.Args) == 1 {
		fmt.Println("fsutils:")
		for _, name := range command.Keys.Main {
			if command.IsDaemon(name) {
				fmt.Printf("\t%s - daemon\n", name)
			} else {
				fmt.Printf("\t%s\n", name)
			}
		}
	} else if err := command.Main(os.Args[1:]...); err != nil {
		fmt.Fprintln(os.Stderr, "fsutils:", err)
		os.Exit(1)
	}
}
