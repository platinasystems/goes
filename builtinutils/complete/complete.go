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

type complete struct{}

func New() complete { return complete{} }

func (complete) String() string { return "-complete" }
func (complete) Tag() string    { return "builtin" }
func (complete) Usage() string  { return "-complete COMMAND [ARGS]..." }

func (complete complete) Main(args ...string) error {
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
		args = append(args[:1], append([]string{"-complete"},
			args[1:]...)...)
		return command.Main(args...)
	}
	return nil
}
