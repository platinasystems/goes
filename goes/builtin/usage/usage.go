// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package usage

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/goes"
)

const Naem = "usage"

type cmd goes.ByName

func New() *cmd { return new(cmd) }

func (*cmd) String() string { return "usage" }
func (*cmd) Tag() string    { return "builtin" }
func (*cmd) Usage() string  { return "usage  COMMAND...\nCOMMAND -usage" }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Complete(args ...string) []string {
	return goes.ByName(*c).Complete(args...)
}

func (c *cmd) Main(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("COMMAND: missing")
	}
	for _, arg := range args {
		g := goes.ByName(*c)[arg]
		if g == nil {
			return fmt.Errorf("%s: not found", arg)
		}
		if strings.IndexRune(g.Usage, '\n') >= 0 {
			fmt.Print("usage:\t",
				strings.Replace(g.Usage, "\n", "\n\t", -1),
				"\n")
		} else {
			fmt.Println("usage:", g.Usage)
		}
	}
	return nil
}

func (*cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print a command synopsis",
	}
}
