// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build amd64 arm

// This is an example goes machine run as daemons w/in another distro.
package main

import (
	"net"

	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/builtin"
	"github.com/platinasystems/go/goes/core"
	"github.com/platinasystems/go/goes/fs"
	"github.com/platinasystems/go/goes/kernel"
	"github.com/platinasystems/go/goes/machine"
	"github.com/platinasystems/go/goes/machine/machined"
	netcmds "github.com/platinasystems/go/goes/net"
	"github.com/platinasystems/go/goes/net/telnetd"
	"github.com/platinasystems/go/goes/redis"
	"github.com/platinasystems/go/goes/test"
	"github.com/platinasystems/go/info/cmdline"
	"github.com/platinasystems/go/info/hostname"
	name "github.com/platinasystems/go/info/machine"
	"github.com/platinasystems/go/info/netlink"
	"github.com/platinasystems/go/info/tests"
	"github.com/platinasystems/go/info/uptime"
	"github.com/platinasystems/go/info/version"
)

func main() {
	command.Plot(builtin.New()...)
	command.Plot(core.New()...)
	command.Plot(fs.New()...)
	command.Plot(kernel.New()...)
	command.Plot(machine.New()...)
	command.Plot(netcmds.New()...)
	command.Plot(redis.New()...)
	command.Plot(telnetd.New())
	command.Plot(test.New()...)
	command.Sort()
	machined.Hook = func() error {
		machined.Plot(
			cmdline.New(),
			hostname.New(),
			name.New("example"),
			netlink.New(),
			uptime.New(),
			version.New(),
		)
		machined.Plot(tests.New()...)
		itfs, err := net.Interfaces()
		if err != nil {
			return nil
		}
		prefixes := make([]string, 0, len(itfs))
		for _, itf := range itfs {
			prefixes = append(prefixes, itf.Name+".")
		}
		machined.Info["netlink"].Prefixes(prefixes...)
		return nil
	}
	goes.Main()
}
