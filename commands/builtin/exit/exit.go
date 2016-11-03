// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package exit

import (
	"os"
	"strconv"
)

type exit struct{}

func New() exit { return exit{} }

func (exit) String() string { return "exit" }
func (exit) Tag() string    { return "builtin" }
func (exit) Usage() string  { return "exit [N]" }

func (exit) Main(args ...string) error {
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

func (exit) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "exit the shell",
	}
}

func (exit) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	exit - exit the shell

SYNOPSIS
	exit [N]

DESCRIPTION
	Exit the shell, returning a status of N, if given, or 0 otherwise.`,
	}
}
