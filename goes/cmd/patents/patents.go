// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package patents

import (
	"fmt"

	"github.com/platinasystems/go/copyright"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "patents"
	Apropos = "print machine patent rights"
	Usage   = "patents"
)

// Some machines may have additional patent claims.
var Others []Other

type Other struct {
	Name, Text string
}

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
	prettyprint("github.com/platinasystems/go", copyright.Patents)
	for _, l := range Others {
		fmt.Print("\n\n")
		prettyprint(l.Name, l.Text)
	}
	return nil
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name }

func prettyprint(title, text string) {
	fmt.Println(title)
	for _ = range title {
		fmt.Print("=")
	}
	fmt.Print("\n\n", text, "\n")
}

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
