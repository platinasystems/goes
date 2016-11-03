// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package cd

import (
	"fmt"
	"os"
)

type cd struct {
	last string
}

func New() *cd { return &cd{} }

func (*cd) String() string { return "cd" }
func (*cd) Tag() string    { return "builtin" }
func (*cd) Usage() string  { return "cd [- | DIRECTORY]" }

func (cd *cd) Main(args ...string) error {
	var dir string

	if len(args) > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	t, err := os.Getwd()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		dir = os.Getenv("HOME")
		if len(dir) == 0 {
			dir = "/root"
		}
	} else if args[0] == "-" {
		if len(cd.last) > 0 {
			dir = cd.last
		}
	} else {
		dir = args[0]
	}
	if len(dir) > 0 {
		err := os.Chdir(dir)
		if err == nil {
			cd.last = t
		}
		return err
	}
	return nil
}

func (*cd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "change the current directory",
	}
}

func (*cd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	cd - change the current directory

SYNOPSIS
	cd [- | DIRECTORY]

DESCRIPTION
	Change the working directory to the given name or the last one if '-'.`,
	}
}
