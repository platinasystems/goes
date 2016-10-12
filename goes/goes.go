// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build linux

// Package goes, combined with a compatibly configured Linux kernel, provides a
// monolithic embedded system.
package goes

import (
	"os"
	"unicode/utf8"

	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/unprompted"
)

func Main(args ...string) error {
	if command.Find(args[0]) != nil {
		return command.Main(args...)
	} else if len(args) > 1 {
		if args[0] == command.Prog {
			script, err := os.Open(args[1])
			if err == nil {
				defer script.Close()
				buf := make([]byte, 4096)
				n, err := script.Read(buf[:])
				if err == nil && utf8.Valid(buf[:n]) {
					script.Seek(0, 0)
					gl := unprompted.New(script).GetLine
					return command.Shell(gl)
				}
			}
		}
		return command.Main(args[1:]...)
	}
	return command.Main("cli")
}
