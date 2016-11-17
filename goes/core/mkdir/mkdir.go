// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package mkdir

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/platinasystems/go/flags"
	"github.com/platinasystems/go/parms"
)

const Name = "mkdir"
const DefaultMode = 0755

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [OPTION]... DIRECTORY..." }

func (cmd) Main(args ...string) error {
	var perm os.FileMode = DefaultMode

	flag, args := flags.New(args, "-p", "-v")
	parm, args := parms.New(args, "-m")

	if len(parm["-m"]) == 0 {
		parm["-m"] = "0755"
	}

	mode, err := strconv.ParseUint(parm["-m"], 8, 64)
	if err != nil {
		return err
	}

	if mode != DefaultMode {
		old := syscall.Umask(0)
		defer syscall.Umask(old)
		perm = os.FileMode(mode)
	}

	f := os.Mkdir
	if flag["-p"] {
		f = os.MkdirAll
	}

	for _, dn := range args {
		if err := f(dn, perm); err != nil {
			return err
		}
		if flag["-v"] {
			fmt.Printf("mkdir: created directory ‘%s’\n", dn)
		}
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "make directories",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	mkdir - make directories

SYNOPSIS
	mkdir [OPTION]... DIRECTORY...

DESCRIPTION
	Create the DIRECTORY(ies), if they do not already exist.

OPTIONS
	-m MODE
		set mode (as in chmod) to the given octal value;
		default, 0777 & umask()

	-p
		don't print error if existing and make parent directories
		as needed

	-v
		print a message for each created directory`,
	}
}
