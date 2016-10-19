// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build linux

// Package goes, combined with a compatibly configured Linux kernel, provides a
// monolithic embedded system.
package goes

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/log"
)

var Exit = os.Exit

func Main() {
	args := os.Args
	if len(args) == 0 {
		return
	}
	if command.Find(args[0]) == nil {
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
	isdaemon := command.IsDaemon(args[0])
	err := command.Main(args...)
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		if isdaemon {
			log.Print("daemon", "err", err)
		} else {
			fmt.Fprintf(os.Stderr, "%s: %v\n",
				filepath.Base(command.Prog()), err)
		}
		Exit(1)
	}
}
