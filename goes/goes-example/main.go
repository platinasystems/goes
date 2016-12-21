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
	"github.com/platinasystems/go/goes/net"
	"github.com/platinasystems/go/goes/net/nld"
	"github.com/platinasystems/go/goes/net/telnetd"
	"github.com/platinasystems/go/goes/redis"
	"github.com/platinasystems/go/goes/redis/redisd"
	"github.com/platinasystems/go/goes/test"
)

func main() {
	g := make(goes.ByName)
	g.Plot(builtin.New()...)
	g.Plot(core.New()...)
	g.Plot(fs.New()...)
	g.Plot(kernel.New()...)
	g.Plot(machine.New()...)
	g.Plot(net.New()...)
	g.Plot(redis.New()...)
	g.Plot(telnetd.New())
	g.Plot(test.New()...)
	redisd.Machine = "example"
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
	g.Main()
}
