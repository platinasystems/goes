// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package complete provides a command that may be used for bash completion
// like this.
//
//	_goes() {
//		COMPREPLY=($(goes -complete ${COMP_WORDS[@]}))
//		return 0
//	}
//	complete -F _goes goes
package complete

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/command"
)

const Name = "-complete"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Tag() string    { return "builtin" }
func (cmd) Usage() string  { return Name + " COMMAND [ARGS]..." }

func (cmd) Main(args ...string) error {
	if len(args) > 0 && args[0] == "goes" {
		args = args[1:]
	}
	switch len(args) {
	case 0:
		for _, name := range command.Keys.Main {
			fmt.Println(name)
		}
	case 1:
		for _, name := range command.Keys.Main {
			if strings.HasPrefix(name, args[0]) {
				fmt.Println(name)
			}
		}
	default:
		args = append(args[:1], append([]string{Name}, args[1:]...)...)
		return command.Main(args...)
	}
	return nil
}
