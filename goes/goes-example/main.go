// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build amd64 arm

// This is an example goes machine run as daemons w/in another distro.
package main

import (
	"bytes"
	"net"
	"os"

	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/commands/builtin"
	"github.com/platinasystems/go/commands/core"
	"github.com/platinasystems/go/commands/fs"
	"github.com/platinasystems/go/commands/kernel"
	"github.com/platinasystems/go/commands/machine"
	"github.com/platinasystems/go/commands/machine/machined"
	"github.com/platinasystems/go/commands/machine/start"
	netcmds "github.com/platinasystems/go/commands/net"
	"github.com/platinasystems/go/commands/net/telnetd"
	"github.com/platinasystems/go/commands/redis"
	"github.com/platinasystems/go/commands/test"
	"github.com/platinasystems/go/goes"
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
	start.Hook = func() error {
		if len(os.Getenv("REDISD")) == 0 {
			return nil
		}
		itfs, err := net.Interfaces()
		if err != nil {
			return nil
		}
		buf := new(bytes.Buffer)
		for i, itf := range itfs {
			if i > 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(itf.Name)
		}
		return os.Setenv("REDISD", buf.String())
	}
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
