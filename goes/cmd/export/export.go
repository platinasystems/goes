// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package export

import (
	"fmt"
	"os"
	"strings"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "export"
	Apropos = "set process configuration"
	Usage   = "export [NAME[=VALUE]]..."
	Man     = `
DESCRIPTION
	Configure the named process environment parameter.

	If no VALUE is given, NAME is reset.

	If no NAMES are supplied, a list of names of all exported variables
	is printed.`
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
	if len(args) == 0 {
		for _, nv := range os.Environ() {
			fmt.Println(nv)
		}
		return nil
	}
	for _, arg := range args {
		eq := strings.Index(arg, "=")
		if eq < 0 {
			if err := os.Unsetenv(arg); err != nil {
				return err
			}
		} else if err := os.Setenv(arg[:eq], arg[eq+1:]); err != nil {
			return err
		}
	}
	return nil
}

func (Command) Man() lang.Alt  { return man }
func (Command) String() string { return Name }
func (Command) Usage() string  { return Usage }
