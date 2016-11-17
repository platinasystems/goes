// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package help

import (
	"fmt"

	"github.com/platinasystems/go/command"
)

const Name = "help"

type help struct{}

func New() help { return help{} }

func (help) String() string { return Name }
func (help) Tag() string    { return "builtin" }
func (help) Usage() string  { return Name + " [COMMAND [ARGS]...]" }

func (a help) Main(args ...string) error {
	if len(args) == 0 {
		for _, k := range command.Keys.Apropos {
			format := "%-15s %s\n"
			if len(k) >= 16 {
				format = "%s\n\t\t%s\n"
			}
			fmt.Printf(format, k, command.Apropos[k])
		}
	} else {
		cmd, err := command.Find(args[0])
		if err != nil {
			return err
		}
		if method, found := cmd.(command.Helper); found {
			fmt.Println(method.Help(args...))
		} else {
			fmt.Println(command.Usage[args[0]])
		}
	}
	return nil
}

func (help) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print command guidance",
	}
}

func (help) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	help - print a command guidance

SYNOPSIS
	help [COMMAND [ARGS]...]

DESCRIPTION
	Print context sensitive command help, if available; otherwise, print
	its usage page.

	Print all available apropos if no COMMAND is given.`,
	}
}
