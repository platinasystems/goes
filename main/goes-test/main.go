// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build amd64 arm

// This is the example goes machine with additional tests.
package main

import (
	"fmt"
	"net"
	"os"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/optional/test/gohellod"
	"github.com/platinasystems/go/internal/optional/test/gopanic"
	"github.com/platinasystems/go/internal/optional/test/gopanicd"
	"github.com/platinasystems/go/internal/optional/test/hellod"
	"github.com/platinasystems/go/internal/optional/test/panic"
	"github.com/platinasystems/go/internal/optional/test/panicd"
	"github.com/platinasystems/go/internal/optional/test/sleeper"
	"github.com/platinasystems/go/internal/optional/test/stringd"
	"github.com/platinasystems/go/internal/required"
	"github.com/platinasystems/go/internal/required/nld"
	"github.com/platinasystems/go/internal/required/redisd"
)

func main() {
	g := make(goes.ByName)
	g.Plot(required.New()...)
	g.Plot(
		gohellod.New(),
		gopanic.New(),
		gopanicd.New(),
		hellod.New(),
		panic.New(),
		panicd.New(),
		sleeper.New(),
		stringd.New(),
	)
	redisd.Machine = "test"
	nld.Hook = func() error {
		itfs, err := net.Interfaces()
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
	if err := g.Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
