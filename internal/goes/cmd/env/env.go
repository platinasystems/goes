// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/goes/lang"
)

const (
	Name    = "env"
	Apropos = "run a program in a modified environment"
	Usage   = "env [NAME[=VALUE... COMMAND [ARGS...]]]"
	Man     = `
DESCRIPTION
	Running 'env' without any arguments prints all environment
	variables.  Runnung 'env' with one argument prints the value of
	the named variable.  Running this with at least one NAME=VALUE
	argument sets each NAME to VALUE in the environment and runs
	COMMAND.`
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
	switch len(args) {
	case 0:
		for _, env := range os.Environ() {
			fmt.Println(env)
		}
	case 1:
		fmt.Println(os.Getenv(args[0]))
	default:
		for {
			eq := strings.Index(args[0], "=")
			if eq < 0 {
				break
			}
			os.Setenv(args[0][:eq], args[0][eq+1:])
			args = args[1:]
		}
		return goes.ByName(*c).Main(args...)
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
