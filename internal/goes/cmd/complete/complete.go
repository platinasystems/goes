// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package complete

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/goes/lang"
)

const (
	Name    = "complete"
	Apropos = "tab to complete command argument"
	Usage   = `
	complete COMMAND [ARGS]...
	COMMAND -complete [ARGS]...`
	Man = `
DESCRIPTION
	This may be used for bash completion of goes commands like this.

	_goes() {
		COMPREPLY=($(goes complete ${COMP_WORDS[@]}))
		return 0
	}
	complete -F _goes goes`
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

func (*cmd) Kind() goes.Kind { return goes.DontFork }

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
