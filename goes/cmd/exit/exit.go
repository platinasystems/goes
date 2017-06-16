// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package exit

import (
	"os"
	"strconv"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "exit"
	Apropos = "exit the shell"
	Usage   = "exit [N]"
	Man     = `
DESCRIPTION
	Exit the shell, returning a status of N, if given, or 0 otherwise.`
)

type Interface interface {
	Apropos() lang.Alt
	Kind() goes.Kind
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }
func (cmd) Kind() goes.Kind   { return goes.DontFork | goes.CantPipe }

func (cmd) Main(args ...string) error {
	var ecode int
	if len(args) != 0 {
		i64, err := strconv.ParseInt(args[0], 0, 0)
		if err != nil {
			return err
		}
		ecode = int(i64)
	}
	os.Exit(ecode)
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
