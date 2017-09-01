// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/helpers"
	"github.com/platinasystems/go/goes/cmd/install"
	"github.com/platinasystems/go/goes/cmd/show_commands"
	"github.com/platinasystems/go/goes/cmd/show_packages"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "goes-installer"
	Apropos = "a self extracting goes machine"
)

func Goes() *goes.Goes {
	g := goes.New(Name, "",
		lang.Alt{
			lang.EnUS: Apropos,
		},
		lang.Alt{})
	g.Plot(helpers.New()...)
	g.Plot(
		install.New(),
		show_commands.New(),
		show_packages.New("license"),
		show_packages.New("version"),
	)
	return g
}
