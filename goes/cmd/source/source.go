// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package source

import (
	"fmt"

	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "source"
	Apropos = "import command script"
	Usage   = "source [-x] FILE"
	Man     = `
DESCRIPTION
	This is equivalent to 'cli [-x] URL'.`
)

type Interface interface {
	Apropos() lang.Alt
	ByName(goes.ByName)
	Kind() goes.Kind
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return new(cmd) }

type cmd goes.ByName

func (*cmd) Apropos() lang.Alt { return apropos }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (*cmd) Kind() goes.Kind { return goes.DontFork | goes.CantPipe }

func (c *cmd) Main(args ...string) error {
	flag, args := flags.New(args, "-x")
	if len(args) == 0 {
		return fmt.Errorf("FILE: missing")
	}
	if len(args) > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	if flag["-x"] {
		args = []string{"cli", "-x", args[0]}
	} else {
		args = []string{"cli", args[0]}
	}
	return goes.ByName(*c).Main(args...)
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
