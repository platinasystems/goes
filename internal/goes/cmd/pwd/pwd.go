// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pwd

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/goes/lang"
)

const (
	Name    = "pwd"
	Apropos = "print working directory"
	Usage   = "pwd [-L]"
	Man     = `
DESCRIPTION
	Print the full filename of the process working directory.

	-L  use PWD from environment, even if it contains symlinks;
	    default avoids symlinks

NOTE 
	This may be different than the context directory.`
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }
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

func (cmd) Man() lang.Alt  { return man }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
