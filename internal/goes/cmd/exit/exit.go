// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package exit

import (
	"os"
	"strconv"

	"github.com/platinasystems/go/internal/goes"
)

const Name = "exit"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) Kind() goes.Kind { return goes.DontFork | goes.CantPipe }
func (cmd) String() string  { return Name }
func (cmd) Usage() string   { return Name + " [N]" }

func (cmd) Main(args ...string) error {
	var ecode int
	if len(args) != 0 {
		i64, err := strconv.ParseInt(args[0], 0, 0)
		if err != nil {
			return err
		}
		ecode = int(i64)
	}
	os.Exit(ecode)
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "exit the shell",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	exit - exit the shell

SYNOPSIS
	exit [N]

DESCRIPTION
	Exit the shell, returning a status of N, if given, or 0 otherwise.`,
	}
}
