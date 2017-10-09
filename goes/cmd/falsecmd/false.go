// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package falsecmd

import (
	"fmt"

	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "false"
	Apropos = "Fail regardless of our ability"
	Usage   = "false"
	Man     = `
DESCRIPTION
	Fail, not matter what. This can not happen in the real world.`
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
	return fmt.Errorf("exit status 1")
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
