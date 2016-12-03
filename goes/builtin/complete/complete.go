// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
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
	"path/filepath"
	"strings"

	"github.com/platinasystems/go/goes"
)

const Name = "-complete"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Tag() string    { return "builtin" }
func (cmd) Usage() string  { return Name + " COMMAND [ARGS]..." }

func (cmd) Main(args ...string) error {
	var ss []string
	if len(args) > 0 && strings.HasPrefix(filepath.Base(args[0]), "goes") {
		args = args[1:]
	}
	if len(args) == 0 {
		ss = goes.Keys.Main
	} else if cmd, err := goes.Find(args[0]); err == nil {
		if method, found := cmd.(goes.Completer); found {
			ss = method.Complete(args[1:]...)
		} else if len(args[1:]) > 0 {
			ss, _ = filepath.Glob(args[len(args)-1] + "*")
		} else {
			ss, _ = filepath.Glob("*")
		}
	} else if len(args) == 1 {
		for _, name := range goes.Keys.Main {
			if strings.HasPrefix(name, args[0]) {
				ss = append(ss, name)
			}
		}
	} else {
		return err
	}
	for _, s := range ss {
		fmt.Println(s)
	}
	return nil
}
