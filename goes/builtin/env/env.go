// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/platinasystems/go/goes"
)

const Name = "env"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Tag() string    { return "builtin" }

func (cmd) Usage() string {
	return Name + " [NAME[=VALUE... COMMAND [ARGS...]]]"
}

func (cmd) Main(args ...string) error {
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
		return goes.Main(args...)
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "run a program in a modified environment",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	env - run a program in a modified environment

SYNOPSIS
	env [NAME[=VALUE... COMMAND [ARGS...]]]

DESCRIPTION
	Running 'env' without any arguments prints all environment
	variables.  Runnung 'env' with one argument prints the value of
	the named variable.  Running this with at least one NAME=VALUE
	argument sets each NAME to VALUE in the environment and runs
	COMMAND.`,
	}
}
