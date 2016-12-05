// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package source

import (
	"fmt"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/internal/flags"
)

const Name = "source"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return "source [-x] FILE" }

func (cmd) Main(args ...string) error {
	flag, args := flags.New(args, "-x")
	if len(args) == 0 {
		return fmt.Errorf("FILE: missing")
	}
	if len(args) > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	if flag["-x"] {
		args = []string{"cli", "-x", args[0]}
	} else {
		args = []string{"cli", args[0]}
	}
	return goes.Main(args...)
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "import command script",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	source - import command script

SYNOPSIS
	source [-x] URL

DESCRIPTION
	This is equivalent to 'cli [-x] URL'.`,
	}
}
