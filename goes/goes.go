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
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/unprompted"
)

func Main(args ...string) (err error) {
	defer func() {
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%s: %v\n",
				filepath.Base(command.Prog), err)
		}
	}()
	if command.Find(args[0]) != nil {
		err = command.Main(args...)
	} else if len(args) > 1 {
		if args[0] == command.Prog {
			if script, ierr := os.Open(args[1]); ierr == nil {
				defer script.Close()
				buf := make([]byte, 4096)
				n, ierr := script.Read(buf[:])
				if ierr == nil && utf8.Valid(buf[:n]) {
					script.Seek(0, 0)
					gl := unprompted.New(script).GetLine
					err = command.Shell(gl)
					return
				}
			}
		}
		err = command.Main(args[1:]...)
	} else {
		err = command.Main("cli")
	}
	return
}
