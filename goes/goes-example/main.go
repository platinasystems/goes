// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build amd64 arm

// This is an example goes machine run as daemons w/in another distro.
package main

import (
	"os"

	"github.com/platinasystems/go/builtinutils"
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/coreutils"
	"github.com/platinasystems/go/diagutils/dlv"
	"github.com/platinasystems/go/fsutils"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/initutils"
	"github.com/platinasystems/go/initutils/start"
	"github.com/platinasystems/go/kutils"
	"github.com/platinasystems/go/machined"
	"github.com/platinasystems/go/machined/info/cmdline"
	"github.com/platinasystems/go/machined/info/hostname"
	"github.com/platinasystems/go/machined/info/netlink"
	"github.com/platinasystems/go/machined/info/tests"
	"github.com/platinasystems/go/machined/info/uptime"
	"github.com/platinasystems/go/machined/info/version"
	"github.com/platinasystems/go/netutils"
	"github.com/platinasystems/go/netutils/telnetd"
	"github.com/platinasystems/go/redisutils"
	"github.com/platinasystems/go/testutils"
)

func main() {
	command.Plot(builtinutils.New()...)
	command.Plot(coreutils.New()...)
	command.Plot(dlv.New()...)
	command.Plot(fsutils.New()...)
	command.Plot(initutils.New()...)
	command.Plot(kutils.New()...)
	command.Plot(machined.New())
	command.Plot(netutils.New()...)
	command.Plot(redisutils.New()...)
	command.Plot(telnetd.New())
	command.Plot(testutils.New()...)
	command.Sort()
	start.Hook = func() error {
		os.Setenv("REDISD", "lo eth0")
		return nil
	}
	machined.Hook = func() error {
		machined.Plot(
			cmdline.New(),
			hostname.New(),
			netlink.New(),
			uptime.New(),
			version.New(),
		)
		machined.Plot(tests.New()...)
		machined.Info["netlink"].Prefixes("lo.", "eth0.")
		return nil
	}
	goes.Main()
}
