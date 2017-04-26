// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package usage

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/goes/lang"
)

const (
	Name    = "usage"
	Apropos = "print a command synopsis"
	Usage   = `
	usage COMMAND...
	COMMAND -usage`
)

type Interface interface {
	Apropos() lang.Alt
	ByName(goes.ByName)
	Complete(...string) []string
	Kind() goes.Kind
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return new(cmd) }

type cmd goes.ByName

func (*cmd) Apropos() lang.Alt { return apropos }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Complete(args ...string) []string {
	var prefix string
	if len(args) > 0 {
		prefix = args[len(args)-1]
	}
	return goes.ByName(*c).Complete(prefix)
}

func (*cmd) Kind() goes.Kind { return goes.DontFork }

func (c *cmd) Main(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("COMMAND: missing")
	}
	for _, arg := range args {
		g := goes.ByName(*c)[arg]
		if g == nil {
			return fmt.Errorf("%s: not found", arg)
		}
		fmt.Print("usage:")
		if !strings.HasPrefix(g.Usage, "\t") {
			fmt.Print("\t")
		}
		fmt.Print(g.Usage)
		if !strings.HasSuffix(g.Usage, "\n") {
			fmt.Println()
		}
	}
	return nil
}

func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
