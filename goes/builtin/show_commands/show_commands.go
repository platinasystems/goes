// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show_commands

import (
	"fmt"

	"github.com/platinasystems/go/goes"
)

const Name = "show-commands"

type cmd goes.ByName

func New() *cmd { return new(cmd) }

func (*cmd) String() string { return "show-commands" }
func (*cmd) Tag() string    { return "builtin" }
func (*cmd) Usage() string  { return "show-commands" }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Main(args ...string) error {
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	for _, name := range goes.ByName(*c).Keys() {
		if g := goes.ByName(*c)[name]; g.Kind == goes.Daemon {
			fmt.Printf("\t%s - daemon\n", name)
		} else {
			fmt.Printf("\t%s\n", name)
		}
	}
	return nil
}

func (*cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "list all commands and daemons",
	}
}
