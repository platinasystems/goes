// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pwd

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/goes/internal/flags"
)

const Name = "pwd"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [-L]" }

func (cmd) Main(args ...string) error {
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

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print working directory",
	}
}

func (cmd) Man() map[string]string {
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
