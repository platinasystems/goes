// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package complete provides a command that may be used for bash completion
// like this.
//
//	_goes() {
//		COMPREPLY=($(goes complete ${COMP_WORDS[@]}))
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

const Name = "complete"

type cmd goes.ByName

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.Builtin }
func (*cmd) String() string  { return Name }

func (*cmd) Usage() string {
	return "complete COMMAND [ARGS]...\nCOMMAND -complete [ARGS]..."
}

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Main(args ...string) error {
	var ss []string
	byName := goes.ByName(*c)
	if len(args) > 0 && strings.HasPrefix(filepath.Base(args[0]), "goes") {
		args = args[1:]
	}
	if len(args) == 0 {
		ss = byName.Complete("")
	} else if g := byName[args[0]]; g != nil {
		if g.Complete != nil {
			ss = g.Complete(args[1:]...)
		} else if len(args[1:]) > 0 {
			ss, _ = filepath.Glob(args[len(args)-1] + "*")
		} else {
			ss, _ = filepath.Glob("*")
		}
	} else if len(args) == 1 {
		ss = byName.Complete(args[0])
	} else {
		return fmt.Errorf("%s: not found", args[0])
	}
	for _, s := range ss {
		fmt.Println(s)
	}
	return nil
}

func (*cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "tab to complete command argument",
	}
}
