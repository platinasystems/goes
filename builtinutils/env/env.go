// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/platinasystems/go/command"
)

type env struct{}

func New() env { return env{} }

func (env) String() string { return "env" }
func (env) Tag() string    { return "builtin" }

func (env) Usage() string {
	return "env [NAME[=VALUE... COMMAND [ARGS...]]]"
}

func (env) Main(args ...string) error {
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
		return command.Main(args...)
	}
	return nil
}

func (env) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "run a program in a modified environment",
	}
}

func (env) Man() map[string]string {
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
