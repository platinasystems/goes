// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cd

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "cd"
	Apropos = "change the current directory"
	Usage   = "cd [- | DIRECTORY]"
	Man     = `
DESCRIPTION
	Change the working directory to the given name or the last one if '-'.`
)

type Interface interface {
	Apropos() lang.Alt
	Kind() goes.Kind
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return &cmd{} }

type cmd struct {
	last string
}

func (*cmd) Apropos() lang.Alt { return apropos }
func (*cmd) Kind() goes.Kind   { return goes.DontFork | goes.CantPipe }

func (cd *cmd) Main(args ...string) error {
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

func (*cmd) Man() lang.Alt  { return man }
func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
