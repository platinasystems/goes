// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package patents

import (
	"fmt"
	"strings"

	. "github.com/platinasystems/go"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "patents"
	Apropos = "print machine patent rights"
	Usage   = "patents"
)

var Packages = func() []map[string]string { return []map[string]string{} }

type Interface interface {
	Apropos() lang.Alt
	Kind() goes.Kind
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }
func (cmd) Kind() goes.Kind   { return goes.DontFork }

func (cmd) Main(args ...string) error {
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	for _, m := range append([]map[string]string{Package}, Packages()...) {
		if patents, found := m["patents"]; found {
			fmt.Print(m["importpath"], ":\n    ",
				strings.Replace(patents, "\n", "\n    ", -1),
				"\n")
		}
	}
	return nil
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
