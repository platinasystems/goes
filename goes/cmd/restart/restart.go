// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package restart

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "restart"
	Apropos = "stop, then start this goes machine"
	Usage   = "restart [STOP, STOP, and REDISD OPTIONS]..."
	Man     = `
DESCRIPTION
	Run the goes machine stop then start commands.

SEE ALSO
	start, stop, and redisd`
)

type Interface interface {
	Apropos() lang.Alt
	ByName(goes.ByName)
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return new(cmd) }

type cmd goes.ByName

func (*cmd) Apropos() lang.Alt { return apropos }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Main(args ...string) error {
	byName := goes.ByName(*c)
	err := byName.Main(append([]string{"stop"}, args...)...)
	if err != nil {
		return err
	}
	return byName.Main(append([]string{"start"}, args...)...)
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
