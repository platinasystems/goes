// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build amd64 arm

// This is an example goes daemon for a debian system.
package main

import (
	"io"
	"os"

	"github.com/platinasystems/go/builtinutils"
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/coreutils"
	"github.com/platinasystems/go/diagutils/dlv"
	"github.com/platinasystems/go/fsutils"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/internal/example"
	"github.com/platinasystems/go/initutils/goesd"
	"github.com/platinasystems/go/kutils"
	"github.com/platinasystems/go/machined"
	"github.com/platinasystems/go/netutils"
	"github.com/platinasystems/go/netutils/telnetd"
	"github.com/platinasystems/go/redisutils"
)

func main() {
	command.Plot(builtinutils.New()...)
	command.Plot(coreutils.New()...)
	command.Plot(dlv.New()...)
	command.Plot(fsutils.New()...)
	command.Plot(goesd.New())
	command.Plot(kutils.New()...)
	command.Plot(netutils.New()...)
	command.Plot(redisutils.New()...)
	command.Plot(telnetd.New())
	command.Sort()
	machined.Hook = example.Machine
	os.Setenv("REDISD_DEVS", "lo eth0")
	err := goes.Main(os.Args...)
	if err != nil && err != io.EOF {
		os.Exit(1)
	}
}
