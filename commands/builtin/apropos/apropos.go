// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package apropos

import (
	"fmt"

	"github.com/platinasystems/go/command"
)

type apropos struct{}

func New() apropos { return apropos{} }

func (apropos) String() string { return "apropos" }
func (apropos) Tag() string    { return "builtin" }
func (apropos) Usage() string  { return "apropos COMMAND..." }

func (a apropos) Main(args ...string) error {
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

func (apropos) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print a short command description",
	}
}

func (apropos) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	apropos - print a short command description

SYNOPSIS
	apropos [COMMAND]...

DESCRIPTION
	Print a short description of given or all COMMANDS.`,
	}
}
