// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cmdline

import (
	"fmt"

	"github.com/platinasystems/go/internal/cmdline"
)

const Name = "cmdline"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [NAME]..." }

func (cmd) Main(args ...string) error {
	keys, m, err := cmdline.New()
	if err != nil {
		return err
	}
	if len(args) > 0 {
		keys = args
	}
	for _, k := range keys {
		if _, found := m[k]; !found {
			return fmt.Errorf("%s: not found", k)
		}
	}
	if len(keys) == 1 {
		fmt.Println(m[keys[0]])
	} else {
		for _, k := range keys {
			fmt.Print(k, ": ", m[k], "\n")
		}
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "parse and print /proc/cmdline variables",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	cmdline - parse and print /proc/cmdline variables

SYNOPSIS
	cmdline [NAME]...

DESCRIPTION
	Print the named or all variables parsed from /proc/cmdline.`,
	}
}
