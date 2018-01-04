// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/cmd/install"
	"github.com/platinasystems/go/goes/lang"
)

var Goes = &goes.Goes{
	NAME: "goes-installer",
	APROPOS: lang.Alt{
		lang.EnUS: "a self extracting goes machine",
	},
	ByName: map[string]cmd.Cmd{
		"install": &install.Command{},
		"show": &goes.Goes{
			NAME:  "show",
			USAGE: "show OBJECT",
			APROPOS: lang.Alt{
				lang.EnUS: "print stuff",
			},
			ByName: map[string]cmd.Cmd{
				"packages": goes.ShowPackages{},
			},
		},
	},
}
