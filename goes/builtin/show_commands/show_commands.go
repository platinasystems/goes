// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show_commands

import (
	"fmt"

	"github.com/platinasystems/go/goes"
)

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return "show-commands" }
func (cmd) Tag() string    { return "builtin" }
func (cmd) Usage() string  { return "show-commands" }

func (cmd) Main(args ...string) error {
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	for _, name := range goes.Keys.Main {
		if goes.IsDaemon(name) {
			fmt.Printf("\t%s - daemon\n", name)
		} else {
			fmt.Printf("\t%s\n", name)
		}
	}
	return nil
}
