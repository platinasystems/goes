// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package pwd

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/flags"
)

type pwd struct{}

func New() pwd { return pwd{} }

func (pwd) String() string { return "pwd" }
func (pwd) Usage() string  { return "pwd [-L]" }

func (pwd) Main(args ...string) error {
	flag, args := flags.New(args, "-L")
	if len(args) != 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if flag["-L"] {
		return fmt.Errorf("FIXME")
	} else {
		fmt.Println(wd)
	}
	return nil
}

func (pwd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print working directory",
	}
}

func (pwd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	pwd - print working directory

SYNOPSIS
	pwd [-L]

DESCRIPTION
	Print the full filename of the process working directory.

	-L  use PWD from environment, even if it contains symlinks;
	    default avoids symlinks

NOTE 
	This may be different than the context directory.`,
	}
}
