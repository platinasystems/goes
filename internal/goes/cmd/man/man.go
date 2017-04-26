// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package man

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/goes/lang"
)

const (
	Name    = "man"
	Apropos = "print command documentation"
	Usage   = `
	man COMMAND...
	COMMAND -man`
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
	n := len(args)
	if n == 0 {
		return fmt.Errorf("COMMAND: missing")
	}
	for i, arg := range args {
		g := goes.ByName(*c)[arg]
		if g == nil {
			return fmt.Errorf("%s: not found", arg)
		}
		man := g.Man.String()
		if len(man) == 0 || strings.HasPrefix(man, "\n") {
			usage := g.Usage
			if strings.HasPrefix(usage, "\t") {
				usage = usage[1:]
			}
			fmt.Print(section.name, "\n\t", g.Name, " - ",
				g.Apropos, "\n\n", section.synopsis, "\n\t",
				usage, "\n")
			if len(man) > 0 {
				if !strings.HasPrefix(man, "\n") {
					fmt.Println()
				}
				fmt.Print(man)
				if !strings.HasSuffix(man, "\n") {
					fmt.Println()
				}
			}
		} else {
			fmt.Println(man)
		}
		if n > 1 && i < n-1 {
			fmt.Println()
		}
	}
	return nil
}

func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	section = struct {
		name, synopsis lang.Alt
	}{
		name: lang.Alt{
			lang.EnUS: "NAME",
		},
		synopsis: lang.Alt{
			lang.EnUS: "SYNOPSIS",
		},
	}
)
