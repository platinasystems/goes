// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build amd64 arm

// This is an example goes machine run as daemons w/in another distro.
package main

import (
	stdnet "net"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/builtin"
	"github.com/platinasystems/go/goes/core"
	"github.com/platinasystems/go/goes/fs"
	"github.com/platinasystems/go/goes/kernel"
	"github.com/platinasystems/go/goes/machine"
	"github.com/platinasystems/go/goes/machine/start"
	"github.com/platinasystems/go/goes/net"
	"github.com/platinasystems/go/goes/net/nld"
	"github.com/platinasystems/go/goes/net/telnetd"
	"github.com/platinasystems/go/goes/redis"
	"github.com/platinasystems/go/goes/test"
)

func main() {
	goes.Plot(builtin.New()...)
	goes.Plot(core.New()...)
	goes.Plot(fs.New()...)
	goes.Plot(kernel.New()...)
	goes.Plot(machine.New()...)
	goes.Plot(net.New()...)
	goes.Plot(redis.New()...)
	goes.Plot(telnetd.New())
	goes.Plot(test.New()...)
	goes.Sort()
	start.Machine = "example"
	nld.Hook = func() error {
		itfs, err := stdnet.Interfaces()
		if err != nil {
			return nil
		}
		prefixes := make([]string, 0, len(itfs))
		for _, itf := range itfs {
			prefixes = append(prefixes, itf.Name+".")
		}
		nld.Prefixes = prefixes
		return nil
	}
	goes.Main()
}
