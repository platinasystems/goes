// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package export

import (
	"fmt"
	"os"
	"strings"

	"github.com/platinasystems/go/internal/goes"
)

const Name = "export"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) Kind() goes.Kind { return goes.DontFork | goes.CantPipe }
func (cmd) String() string  { return Name }
func (cmd) Usage() string   { return Name + " [NAME[=VALUE]]..." }

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

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "set process configuration",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	export - process configuration

SYNOPSIS
	export [NAME=[VALUE]]...

DESCRIPTION
	Configure the named process environment parameter.

	If no VALUE is given, NAME is reset.

	If no NAMES are supplied, a list of names of all exported variables
	is printed.`,
	}
}
