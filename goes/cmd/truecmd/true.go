// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package truecmd

import (
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "true"
	Apropos = "Be successful not matter what"
	Usage   = "true"
	Man     = `
DESCRIPTION
	Be successful no matter what!`
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

func (cmd) Main(_ ...string) error {
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
