// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package main

import (
	"io"
	"os"

	"github.com/platinasystems/go/builtinutils"
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/goes"
)

func main() {
	command.Plot(builtinutils.New()...)
	command.Sort()
	err := goes.Main(os.Args...)
	if err != nil && err != io.EOF {
		os.Exit(1)
	}
}
