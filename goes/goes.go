// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build linux

// Package goes, combined with a compatibly configured Linux kernel, provides a
// monolithic embedded system.
package goes

import (
	"io/ioutil"
	"os"
	"unicode/utf8"

	"github.com/platinasystems/go/command"
)

var Exit = os.Exit

func Main() {
	args := os.Args
	if len(args) == 0 {
		return
	}
	if _, err := command.Find(args[0]); err != nil {
		if args[0] == "/usr/bin/goes" && len(args) > 2 {
			buf, err := ioutil.ReadFile(args[1])
			if err == nil && utf8.Valid(buf) {
				args = []string{"source", args[1]}
			} else {
				args = args[1:]
			}
		} else {
			args = args[1:]
		}
	}
	if len(args) == 0 {
		args = []string{"cli"}
	}
	err := command.Main(args...)
	if err != nil {
		Exit(1)
	}
}
