// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package exit

import (
	"os"
	"strconv"

	"github.com/platinasystems/go/goes/cmd"
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

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Kind() cmd.Kind    { return cmd.DontFork | cmd.CantPipe }

func (Command) Main(args ...string) error {
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

func (Command) Man() lang.Alt  { return man }
func (Command) String() string { return Name }
func (Command) Usage() string  { return Usage }
