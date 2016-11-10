// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package apropos

import (
	"fmt"

	"github.com/platinasystems/go/command"
)

const Name = "apropos"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Tag() string    { return "builtin" }
func (cmd) Usage() string  { return Name + " COMMAND..." }

func (cmd) Main(args ...string) error {
	printApropos := func(k, v string) {
		format := "%-15s %s\n"
		if len(k) >= 16 {
			format = "%s\n\t\t%s\n"
		}
		fmt.Printf(format, k, v)
	}
	if len(args) == 0 {
		for _, k := range command.Keys.Apropos {
			printApropos(k, command.Apropos[k])
		}
	} else {
		for _, k := range args {
			v, found := command.Apropos[k]
			if !found {
				v = "has no apropos"
			}
			printApropos(k, v)
		}
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print a short command description",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	apropos - print a short command description

SYNOPSIS
	apropos [COMMAND]...

DESCRIPTION
	Print a short description of given or all COMMANDS.`,
	}
}
