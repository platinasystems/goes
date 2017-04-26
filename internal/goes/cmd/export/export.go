// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package export

import (
	"fmt"
	"os"
	"strings"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/goes/lang"
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
