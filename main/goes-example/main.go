// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build amd64 arm

// This is an example goes machine run as daemons w/in another distro.
package main

import (
	stdnet "net"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/required"
	"github.com/platinasystems/go/internal/required/nld"
	"github.com/platinasystems/go/internal/required/redisd"
)

func main() {
	g := make(goes.ByName)
	g.Plot(required.New()...)
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
